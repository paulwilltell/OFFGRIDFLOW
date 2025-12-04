// Package handlers provides HTTP handlers for the OffGridFlow API.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/example/offgridflow/internal/api/graph"
	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/compliance/csrd"
	"github.com/example/offgridflow/internal/emissions"
	"github.com/example/offgridflow/internal/ingestion"
)

// GraphQLHandlerConfig holds dependencies for the GraphQL handler.
type GraphQLHandlerConfig struct {
	ActivityStore    ingestion.ActivityStore
	Scope1Calculator *emissions.Scope1Calculator
	Scope2Calculator *emissions.Scope2Calculator
	Scope3Calculator *emissions.Scope3Calculator
	CSRDMapper       csrd.CSRDMapper
	Logger           *slog.Logger
}

// GraphQLRequest represents an incoming GraphQL request.
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response.
type GraphQLResponse struct {
	Data   interface{}    `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error.
type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []GraphQLLocation      `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQLLocation represents a location in the GraphQL query.
type GraphQLLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// NewGraphQLHandler creates a new GraphQL HTTP handler.
// This handler provides a simplified GraphQL execution engine that handles
// the core query operations for emissions and compliance data.
func NewGraphQLHandler(cfg GraphQLHandlerConfig) http.HandlerFunc {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default().With("component", "graphql-handler")
	}

	// Create production resolver
	resolver, err := graph.NewProductionQueryResolver(graph.ProductionQueryResolverConfig{
		ActivityStore:    cfg.ActivityStore,
		Scope1Calculator: cfg.Scope1Calculator,
		Scope2Calculator: cfg.Scope2Calculator,
		Scope3Calculator: cfg.Scope3Calculator,
		CSRDMapper:       cfg.CSRDMapper,
		Logger:           logger,
	})
	if err != nil {
		logger.Error("failed to create production resolver", "error", err)
		// Fall back to default resolver
		resolver = nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Only accept POST for GraphQL
		if r.Method != http.MethodPost {
			responders.MethodNotAllowed(w, http.MethodPost)
			return
		}

		// Parse request
		var gqlReq GraphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&gqlReq); err != nil {
			respondGraphQLError(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		if gqlReq.Query == "" {
			respondGraphQLError(w, "Query is required", http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		// Simple query parser - handles basic queries
		// In production, use a full GraphQL library like gqlgen or graphql-go
		result, err := executeQuery(ctx, gqlReq, resolver, logger)
		if err != nil {
			respondGraphQLError(w, err.Error(), http.StatusOK) // GraphQL errors return 200
			return
		}

		responders.JSON(w, http.StatusOK, GraphQLResponse{
			Data: result,
		})
	}
}

// executeQuery executes a GraphQL query using the resolver.
func executeQuery(ctx context.Context, req GraphQLRequest, resolver *graph.ProductionQueryResolver, logger *slog.Logger) (interface{}, error) {
	// Simple query matching - in production use a proper GraphQL parser
	query := req.Query

	// Handle introspection for tools like GraphiQL
	if containsSubstring(query, "__schema") || containsSubstring(query, "__type") {
		return handleIntrospection(), nil
	}

	// Handle placeholder query
	if containsSubstring(query, "placeholder") {
		if resolver != nil {
			msg, err := resolver.Placeholder(ctx)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{"placeholder": msg}, nil
		}
		return map[string]interface{}{"placeholder": "GraphQL API is live"}, nil
	}

	// Handle emissions query
	if containsSubstring(query, "emissions") && !containsSubstring(query, "emissionsSummary") {
		if resolver != nil {
			filter := graph.EmissionsFilter{}
			if vars := req.Variables; vars != nil {
				if scope, ok := vars["scope"].(string); ok {
					filter.Scope = &scope
				}
			}
			conn, err := resolver.Emissions(ctx, filter)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{"emissions": conn}, nil
		}
		return map[string]interface{}{"emissions": map[string]interface{}{
			"edges":      []interface{}{},
			"pageInfo":   map[string]interface{}{"hasNextPage": false, "hasPreviousPage": false},
			"totalCount": 0,
		}}, nil
	}

	// Handle emissionsSummary query
	if containsSubstring(query, "emissionsSummary") {
		year := 2024
		if vars := req.Variables; vars != nil {
			if y, ok := vars["year"].(float64); ok {
				year = int(y)
			}
		}
		if resolver != nil {
			summary, err := resolver.EmissionsSummary(ctx, year)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{"emissionsSummary": summary}, nil
		}
		return map[string]interface{}{"emissionsSummary": map[string]interface{}{
			"scope1TonnesCO2e": 0,
			"scope2TonnesCO2e": 0,
			"scope3TonnesCO2e": 0,
			"totalTonnesCO2e":  0,
			"year":             year,
		}}, nil
	}

	// Handle complianceStatus query
	if containsSubstring(query, "complianceStatus") {
		framework := "CSRD"
		if vars := req.Variables; vars != nil {
			if f, ok := vars["framework"].(string); ok {
				framework = f
			}
		}
		if resolver != nil {
			status, err := resolver.ComplianceStatus(ctx, framework)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{"complianceStatus": status}, nil
		}
		return map[string]interface{}{"complianceStatus": map[string]interface{}{
			"framework":        framework,
			"status":           "not_started",
			"requiredMetrics":  []string{},
			"completedMetrics": []string{},
			"missingMetrics":   []string{},
			"score":            0,
		}}, nil
	}

	return nil, fmt.Errorf("unsupported query operation")
}

// handleIntrospection returns a minimal schema introspection response.
func handleIntrospection() map[string]interface{} {
	return map[string]interface{}{
		"__schema": map[string]interface{}{
			"queryType": map[string]interface{}{
				"name": "Query",
			},
			"mutationType": map[string]interface{}{
				"name": "Mutation",
			},
			"types": []map[string]interface{}{
				{"name": "Query", "kind": "OBJECT"},
				{"name": "Mutation", "kind": "OBJECT"},
				{"name": "Emission", "kind": "OBJECT"},
				{"name": "EmissionsConnection", "kind": "OBJECT"},
				{"name": "EmissionsSummary", "kind": "OBJECT"},
				{"name": "ComplianceStatus", "kind": "OBJECT"},
				{"name": "Activity", "kind": "OBJECT"},
			},
		},
	}
}

// respondGraphQLError sends a GraphQL-formatted error response.
func respondGraphQLError(w http.ResponseWriter, message string, status int) {
	resp := GraphQLResponse{
		Errors: []GraphQLError{
			{Message: message},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// containsSubstring checks if a string contains a substring (case-insensitive).
func containsSubstring(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
