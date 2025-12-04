// Package graph provides GraphQL resolvers for the OffGridFlow API.
//
// This package implements a library-agnostic resolver pattern that can be adapted
// to work with any GraphQL library (gqlgen, graphql-go, 99designs/gqlgen, etc.).
//
// Architecture:
//
//	RootResolver
//	├── Query           - Read-only data fetching
//	│   ├── placeholder - Health check endpoint
//	│   ├── emissions   - Emissions data queries
//	│   └── compliance  - Compliance report queries
//	└── Mutation        - Data modification operations
//	    └── (future)    - Activity ingestion, settings, etc.
//
// Usage:
//
//	resolver, err := graph.NewRootResolver(
//	    graph.WithQueryResolver(queryImpl),
//	    graph.WithMutationResolver(mutationImpl),
//	)
package graph

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// -----------------------------------------------------------------------------
// Error Definitions
// -----------------------------------------------------------------------------

var (
	// ErrNilQueryResolver is returned when a RootResolver is created without
	// a QueryResolver implementation.
	ErrNilQueryResolver = errors.New("graph: nil QueryResolver")

	// ErrContextCanceled indicates the request context was canceled.
	ErrContextCanceled = errors.New("graph: context canceled")

	// ErrNotImplemented indicates a resolver method is not yet implemented.
	ErrNotImplemented = errors.New("graph: not implemented")

	// ErrUnauthorized indicates the user is not authorized for the operation.
	ErrUnauthorized = errors.New("graph: unauthorized")

	// ErrNotFound indicates the requested resource was not found.
	ErrNotFound = errors.New("graph: not found")
)

// -----------------------------------------------------------------------------
// Resolver Interfaces
// -----------------------------------------------------------------------------

// QueryResolver defines the contract for resolving fields on the GraphQL
// Query type as defined in schema.graphql.
type QueryResolver interface {
	// Placeholder resolves the "placeholder" field on the Query type.
	// Used for health checks and API smoke tests.
	Placeholder(ctx context.Context) (string, error)

	// Emissions resolves emissions data queries.
	// Returns aggregated emissions data for the authenticated tenant.
	Emissions(ctx context.Context, filter EmissionsFilter) (*EmissionsConnection, error)

	// EmissionsSummary resolves the emissions summary.
	// Returns totals across all scopes.
	EmissionsSummary(ctx context.Context, year int) (*EmissionsSummary, error)

	// ComplianceStatus resolves compliance status for a specific framework.
	ComplianceStatus(ctx context.Context, framework string) (*ComplianceStatus, error)
}

// MutationResolver defines the contract for resolving mutations.
type MutationResolver interface {
	// CreateActivity creates a new activity record for emissions calculation.
	CreateActivity(ctx context.Context, input CreateActivityInput) (*Activity, error)

	// UpdateOrganizationSettings updates organization-level settings.
	UpdateOrganizationSettings(ctx context.Context, input UpdateOrgSettingsInput) (*OrgSettings, error)
}

// SubscriptionResolver defines the contract for GraphQL subscriptions.
type SubscriptionResolver interface {
	// EmissionsUpdated subscribes to real-time emissions calculation updates.
	EmissionsUpdated(ctx context.Context) (<-chan *EmissionsUpdate, error)
}

// -----------------------------------------------------------------------------
// GraphQL Types
// -----------------------------------------------------------------------------

// EmissionsFilter defines filter parameters for emissions queries.
type EmissionsFilter struct {
	Scope     *string    `json:"scope,omitempty"`
	StartDate *time.Time `json:"startDate,omitempty"`
	EndDate   *time.Time `json:"endDate,omitempty"`
	Region    *string    `json:"region,omitempty"`
	First     *int       `json:"first,omitempty"`
	After     *string    `json:"after,omitempty"`
}

// EmissionsConnection represents a paginated list of emissions.
type EmissionsConnection struct {
	Edges      []*EmissionsEdge `json:"edges"`
	PageInfo   *PageInfo        `json:"pageInfo"`
	TotalCount int              `json:"totalCount"`
}

// EmissionsEdge represents an edge in the emissions connection.
type EmissionsEdge struct {
	Node   *Emission `json:"node"`
	Cursor string    `json:"cursor"`
}

// Emission represents a single emission record.
type Emission struct {
	ID                string    `json:"id"`
	Scope             string    `json:"scope"`
	Category          string    `json:"category"`
	Source            string    `json:"source"`
	AmountKgCO2e      float64   `json:"amountKgCO2e"`
	AmountTonnesCO2e  float64   `json:"amountTonnesCO2e"`
	Region            string    `json:"region"`
	PeriodStart       time.Time `json:"periodStart"`
	PeriodEnd         time.Time `json:"periodEnd"`
	CalculatedAt      time.Time `json:"calculatedAt"`
	EmissionFactor    float64   `json:"emissionFactor"`
	EmissionFactorRef string    `json:"emissionFactorRef,omitempty"`
	DataQuality       string    `json:"dataQuality"`
}

// PageInfo provides pagination information.
type PageInfo struct {
	HasNextPage     bool    `json:"hasNextPage"`
	HasPreviousPage bool    `json:"hasPreviousPage"`
	StartCursor     *string `json:"startCursor,omitempty"`
	EndCursor       *string `json:"endCursor,omitempty"`
}

// EmissionsSummary provides aggregated emissions data.
type EmissionsSummary struct {
	Scope1TonnesCO2e float64            `json:"scope1TonnesCO2e"`
	Scope2TonnesCO2e float64            `json:"scope2TonnesCO2e"`
	Scope3TonnesCO2e float64            `json:"scope3TonnesCO2e"`
	TotalTonnesCO2e  float64            `json:"totalTonnesCO2e"`
	Year             int                `json:"year"`
	RegionBreakdown  map[string]float64 `json:"regionBreakdown,omitempty"`
	Timestamp        time.Time          `json:"timestamp"`
}

// ComplianceStatus represents compliance status for a regulatory framework.
type ComplianceStatus struct {
	Framework        string    `json:"framework"`
	Status           string    `json:"status"` // "compliant", "partial", "non_compliant", "not_started"
	RequiredMetrics  []string  `json:"requiredMetrics"`
	CompletedMetrics []string  `json:"completedMetrics"`
	MissingMetrics   []string  `json:"missingMetrics"`
	Score            float64   `json:"score"` // 0-100 percentage
	LastUpdated      time.Time `json:"lastUpdated"`
}

// Activity represents an ingested activity for emissions calculation.
type Activity struct {
	ID          string    `json:"id"`
	Source      string    `json:"source"`
	Category    string    `json:"category"`
	Quantity    float64   `json:"quantity"`
	Unit        string    `json:"unit"`
	Location    string    `json:"location"`
	PeriodStart time.Time `json:"periodStart"`
	PeriodEnd   time.Time `json:"periodEnd"`
	CreatedAt   time.Time `json:"createdAt"`
}

// CreateActivityInput represents input for creating an activity.
type CreateActivityInput struct {
	Source      string    `json:"source"`
	Category    string    `json:"category"`
	Quantity    float64   `json:"quantity"`
	Unit        string    `json:"unit"`
	Location    string    `json:"location"`
	PeriodStart time.Time `json:"periodStart"`
	PeriodEnd   time.Time `json:"periodEnd"`
}

// OrgSettings represents organization settings.
type OrgSettings struct {
	OrgID              string `json:"orgId"`
	DefaultRegion      string `json:"defaultRegion"`
	Methodology        string `json:"methodology"` // "location-based" or "market-based"
	ReportingFramework string `json:"reportingFramework"`
}

// UpdateOrgSettingsInput represents input for updating org settings.
type UpdateOrgSettingsInput struct {
	DefaultRegion      *string `json:"defaultRegion,omitempty"`
	Methodology        *string `json:"methodology,omitempty"`
	ReportingFramework *string `json:"reportingFramework,omitempty"`
}

// EmissionsUpdate represents a real-time emissions update notification.
type EmissionsUpdate struct {
	EventType string    `json:"eventType"` // "created", "updated", "deleted"
	Emission  *Emission `json:"emission"`
	Timestamp time.Time `json:"timestamp"`
}

// -----------------------------------------------------------------------------
// Root Resolver
// -----------------------------------------------------------------------------

// RootResolver is the main entry point for GraphQL resolvers.
//
// In a real GraphQL server, your framework (e.g. gqlgen, graphql-go, etc.)
// will typically expect a struct with methods or fields that it can call to
// resolve queries, mutations, and subscriptions.
//
// RootResolver is intentionally library-agnostic: you can adapt it to whichever
// GraphQL library you choose without changing the core business logic.
type RootResolver struct {
	Query        QueryResolver
	Mutation     MutationResolver
	Subscription SubscriptionResolver
	logger       *slog.Logger
}

// RootResolverOption configures a RootResolver.
type RootResolverOption func(*RootResolver)

// WithQueryResolver sets the query resolver.
func WithQueryResolver(qr QueryResolver) RootResolverOption {
	return func(r *RootResolver) {
		r.Query = qr
	}
}

// WithMutationResolver sets the mutation resolver.
func WithMutationResolver(mr MutationResolver) RootResolverOption {
	return func(r *RootResolver) {
		r.Mutation = mr
	}
}

// WithSubscriptionResolver sets the subscription resolver.
func WithSubscriptionResolver(sr SubscriptionResolver) RootResolverOption {
	return func(r *RootResolver) {
		r.Subscription = sr
	}
}

// WithLogger sets the logger for the resolver.
func WithLogger(logger *slog.Logger) RootResolverOption {
	return func(r *RootResolver) {
		r.logger = logger
	}
}

// NewRootResolver constructs a RootResolver with the provided options.
//
// At minimum, a QueryResolver must be provided. This upfront validation ensures
// you don't accidentally run the server with an incompletely wired resolver graph.
func NewRootResolver(opts ...RootResolverOption) (*RootResolver, error) {
	r := &RootResolver{
		logger: slog.Default().With("component", "graphql-resolver"),
	}

	for _, opt := range opts {
		opt(r)
	}

	if r.Query == nil {
		return nil, ErrNilQueryResolver
	}

	return r, nil
}

// -----------------------------------------------------------------------------
// Default Query Resolver
// -----------------------------------------------------------------------------

// DefaultQueryResolver is a minimal, production-safe implementation of
// QueryResolver. It is primarily useful for bootstrap and smoke tests.
type DefaultQueryResolver struct {
	// Message is the string returned by the placeholder field.
	// If empty, a sensible default message is used.
	Message string
	logger  *slog.Logger
}

// NewDefaultQueryResolver constructs a DefaultQueryResolver with the given
// message. If message is empty, a default message is used instead.
func NewDefaultQueryResolver(message string) *DefaultQueryResolver {
	if message == "" {
		message = "GraphQL API is live. Replace DefaultQueryResolver with your domain-specific implementation."
	}

	return &DefaultQueryResolver{
		Message: message,
		logger:  slog.Default().With("component", "default-query-resolver"),
	}
}

// Placeholder returns the configured message, respecting context cancellation.
func (r *DefaultQueryResolver) Placeholder(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("%w: %v", ErrContextCanceled, err)
	}
	return r.Message, nil
}

// Emissions returns a stub emissions connection.
// Override this in a production resolver to fetch real data.
func (r *DefaultQueryResolver) Emissions(ctx context.Context, filter EmissionsFilter) (*EmissionsConnection, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrContextCanceled, err)
	}

	// Return empty connection for default resolver
	return &EmissionsConnection{
		Edges:      []*EmissionsEdge{},
		PageInfo:   &PageInfo{},
		TotalCount: 0,
	}, nil
}

// EmissionsSummary returns a stub emissions summary.
// Override this in a production resolver to calculate real totals.
func (r *DefaultQueryResolver) EmissionsSummary(ctx context.Context, year int) (*EmissionsSummary, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrContextCanceled, err)
	}

	return &EmissionsSummary{
		Scope1TonnesCO2e: 0,
		Scope2TonnesCO2e: 0,
		Scope3TonnesCO2e: 0,
		TotalTonnesCO2e:  0,
		Year:             year,
		Timestamp:        time.Now(),
	}, nil
}

// ComplianceStatus returns a stub compliance status.
// Override this in a production resolver to check real compliance.
func (r *DefaultQueryResolver) ComplianceStatus(ctx context.Context, framework string) (*ComplianceStatus, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrContextCanceled, err)
	}

	return &ComplianceStatus{
		Framework:        framework,
		Status:           "not_started",
		RequiredMetrics:  []string{},
		CompletedMetrics: []string{},
		MissingMetrics:   []string{},
		Score:            0,
		LastUpdated:      time.Now(),
	}, nil
}
