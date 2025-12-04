// Package allocation provides emissions allocation capabilities.
//
// Allocation is the process of distributing emissions across organizational
// dimensions (business units, sites, products, cost centers) according to
// defined rules and methodologies.
//
// The GHG Protocol allows several allocation approaches:
//   - Physical allocation (based on mass, volume, energy)
//   - Economic allocation (based on revenue, cost)
//   - Activity-based allocation (based on operational metrics)
//
// This package provides a rule-based engine for configuring and executing
// these allocation strategies.
package allocation

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// =============================================================================
// Dimension Constants
// =============================================================================

// Dimension represents the organizational axis along which emissions are allocated.
// This determines how emissions are attributed across the organization.
type Dimension string

const (
	// DimensionBusinessUnit allocates to organizational units/divisions.
	DimensionBusinessUnit Dimension = "BUSINESS_UNIT"

	// DimensionSite allocates to physical locations/facilities.
	DimensionSite Dimension = "SITE"

	// DimensionProduct allocates to products or product lines.
	DimensionProduct Dimension = "PRODUCT"

	// DimensionProject allocates to specific projects.
	DimensionProject Dimension = "PROJECT"

	// DimensionCostCenter allocates to financial cost centers.
	DimensionCostCenter Dimension = "COST_CENTER"

	// DimensionDepartment allocates to departments.
	DimensionDepartment Dimension = "DEPARTMENT"

	// DimensionCustom allows custom allocation dimensions.
	DimensionCustom Dimension = "CUSTOM"
)

// String returns the string representation of the dimension.
func (d Dimension) String() string {
	return string(d)
}

// IsValid returns true if the dimension is a recognized value.
func (d Dimension) IsValid() bool {
	switch d {
	case DimensionBusinessUnit, DimensionSite, DimensionProduct,
		DimensionProject, DimensionCostCenter, DimensionDepartment,
		DimensionCustom:
		return true
	default:
		return false
	}
}

// =============================================================================
// Allocation Method Constants
// =============================================================================

// AllocationMethod specifies how allocation percentages are determined.
type AllocationMethod string

const (
	// MethodFixed uses static percentage allocations.
	MethodFixed AllocationMethod = "FIXED"

	// MethodRevenue allocates proportionally by revenue.
	MethodRevenue AllocationMethod = "REVENUE"

	// MethodHeadcount allocates proportionally by employee count.
	MethodHeadcount AllocationMethod = "HEADCOUNT"

	// MethodArea allocates proportionally by floor space/area.
	MethodArea AllocationMethod = "AREA"

	// MethodEnergy allocates proportionally by energy consumption.
	MethodEnergy AllocationMethod = "ENERGY"

	// MethodProduction allocates proportionally by production volume.
	MethodProduction AllocationMethod = "PRODUCTION"

	// MethodExpression uses a custom expression for allocation.
	MethodExpression AllocationMethod = "EXPRESSION"
)

// String returns the string representation of the method.
func (m AllocationMethod) String() string {
	return string(m)
}

// =============================================================================
// Sentinel Errors
// =============================================================================

var (
	// ErrRuleMissingID is returned when a rule has no ID.
	ErrRuleMissingID = errors.New("allocation: rule ID is required")

	// ErrRuleMissingDimension is returned when a rule has no dimension.
	ErrRuleMissingDimension = errors.New("allocation: rule dimension is required")

	// ErrRuleMissingExpression is returned when a rule has no expression.
	ErrRuleMissingExpression = errors.New("allocation: rule expression is required")

	// ErrInvalidDimension is returned for unrecognized dimensions.
	ErrInvalidDimension = errors.New("allocation: invalid dimension")

	// ErrInvalidPercentage is returned when allocations don't sum to 100%.
	ErrInvalidPercentage = errors.New("allocation: percentages must sum to 100")

	// ErrRuleNotFound is returned when a rule cannot be found.
	ErrRuleNotFound = errors.New("allocation: rule not found")

	// ErrCircularDependency is returned when rules have circular dependencies.
	ErrCircularDependency = errors.New("allocation: circular dependency detected")
)

// =============================================================================
// Rule Model
// =============================================================================

// Rule defines an allocation rule that specifies how emissions should be
// distributed across a dimension.
//
// Rules can use different methods:
//   - Fixed percentages: Specify exact allocation percentages
//   - Driver-based: Use metrics like revenue, headcount, or area
//   - Expression-based: Use a domain-specific expression language
//
// Example fixed allocation:
//
//	rule := Rule{
//	    ID:         "allocate-by-bu",
//	    Dimension:  DimensionBusinessUnit,
//	    Method:     MethodFixed,
//	    Allocations: []AllocationTarget{
//	        {TargetID: "engineering", Percentage: 40.0},
//	        {TargetID: "sales", Percentage: 35.0},
//	        {TargetID: "operations", Percentage: 25.0},
//	    },
//	}
type Rule struct {
	// ID is a unique identifier for this rule.
	ID string `json:"id"`

	// Name is a human-readable name for the rule.
	Name string `json:"name,omitempty"`

	// Description provides context for this rule.
	Description string `json:"description,omitempty"`

	// Dimension is the organizational axis for allocation.
	Dimension Dimension `json:"dimension"`

	// Method specifies how allocation percentages are determined.
	Method AllocationMethod `json:"method"`

	// Expression is the allocation expression (for MethodExpression).
	// Supports a domain-specific syntax for defining allocation logic.
	Expression string `json:"expression,omitempty"`

	// Allocations defines fixed allocation targets and percentages.
	// Used when Method is MethodFixed.
	Allocations []AllocationTarget `json:"allocations,omitempty"`

	// Filters limit which emissions this rule applies to.
	Filters []RuleFilter `json:"filters,omitempty"`

	// Priority determines rule evaluation order (higher = first).
	Priority int `json:"priority,omitempty"`

	// Enabled indicates whether this rule is active.
	Enabled bool `json:"enabled"`

	// OrgID identifies the organization this rule belongs to.
	OrgID string `json:"org_id,omitempty"`

	// CreatedAt is when this rule was created.
	CreatedAt time.Time `json:"created_at,omitempty"`

	// UpdatedAt is when this rule was last modified.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// AllocationTarget specifies a single allocation destination.
type AllocationTarget struct {
	// TargetID identifies the allocation target (e.g., "engineering").
	TargetID string `json:"target_id"`

	// TargetName is a human-readable name.
	TargetName string `json:"target_name,omitempty"`

	// Percentage is the allocation percentage (0-100).
	Percentage float64 `json:"percentage"`

	// DriverField specifies which field to use for driver-based allocation.
	// Used when Rule.Method is revenue, headcount, etc.
	DriverField string `json:"driver_field,omitempty"`
}

// RuleFilter limits which emissions a rule applies to.
type RuleFilter struct {
	// Field is the field to filter on (e.g., "scope", "source", "region").
	Field string `json:"field"`

	// Operator is the comparison operator (eq, neq, in, nin, contains).
	Operator string `json:"operator"`

	// Value is the value to compare against.
	Value string `json:"value"`
}

// Validate performs comprehensive validation on the rule.
func (r Rule) Validate() error {
	var errs []error

	if strings.TrimSpace(r.ID) == "" {
		errs = append(errs, ErrRuleMissingID)
	}

	if !r.Dimension.IsValid() {
		errs = append(errs, fmt.Errorf("%w: %q", ErrInvalidDimension, r.Dimension))
	}

	// Validate based on method
	switch r.Method {
	case MethodFixed:
		if len(r.Allocations) == 0 {
			errs = append(errs, errors.New("allocation: fixed method requires allocations"))
		} else if err := r.validateAllocations(); err != nil {
			errs = append(errs, err)
		}

	case MethodExpression:
		if strings.TrimSpace(r.Expression) == "" {
			errs = append(errs, ErrRuleMissingExpression)
		}

	case MethodRevenue, MethodHeadcount, MethodArea, MethodEnergy, MethodProduction:
		// Driver-based methods need allocation targets with driver fields
		if len(r.Allocations) == 0 {
			errs = append(errs, errors.New("allocation: driver-based method requires targets"))
		}

	default:
		if strings.TrimSpace(string(r.Method)) == "" {
			errs = append(errs, errors.New("allocation: method is required"))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("rule %q validation failed: %w", r.ID, errors.Join(errs...))
	}

	return nil
}

// validateAllocations checks that fixed allocations sum to 100%.
func (r Rule) validateAllocations() error {
	if len(r.Allocations) == 0 {
		return nil
	}

	var total float64
	for _, a := range r.Allocations {
		if a.Percentage < 0 || a.Percentage > 100 {
			return fmt.Errorf("invalid percentage %.2f for target %q", a.Percentage, a.TargetID)
		}
		total += a.Percentage
	}

	// Allow small floating-point tolerance
	if total < 99.99 || total > 100.01 {
		return fmt.Errorf("%w: got %.2f%%", ErrInvalidPercentage, total)
	}

	return nil
}

// IsZero reports whether the rule is effectively uninitialized.
func (r Rule) IsZero() bool {
	return strings.TrimSpace(r.ID) == "" &&
		strings.TrimSpace(string(r.Dimension)) == ""
}

// String returns a concise human-readable representation.
func (r Rule) String() string {
	return fmt.Sprintf("Rule{id=%q, dimension=%s, method=%s, enabled=%t}",
		r.ID, r.Dimension, r.Method, r.Enabled)
}

// Clone creates a deep copy of the rule.
func (r Rule) Clone() Rule {
	clone := r

	// Deep copy allocations
	if r.Allocations != nil {
		clone.Allocations = make([]AllocationTarget, len(r.Allocations))
		copy(clone.Allocations, r.Allocations)
	}

	// Deep copy filters
	if r.Filters != nil {
		clone.Filters = make([]RuleFilter, len(r.Filters))
		copy(clone.Filters, r.Filters)
	}

	return clone
}

// JSON serializes the rule to JSON bytes.
func (r Rule) JSON() ([]byte, error) {
	return json.Marshal(r)
}

// =============================================================================
// Rule Builder
// =============================================================================

// RuleBuilder provides a fluent interface for constructing rules.
type RuleBuilder struct {
	rule Rule
}

// NewRuleBuilder creates a new rule builder.
func NewRuleBuilder(id string) *RuleBuilder {
	return &RuleBuilder{
		rule: Rule{
			ID:        id,
			Enabled:   true,
			CreatedAt: time.Now().UTC(),
		},
	}
}

// WithName sets the rule name.
func (b *RuleBuilder) WithName(name string) *RuleBuilder {
	b.rule.Name = name
	return b
}

// WithDescription sets the rule description.
func (b *RuleBuilder) WithDescription(desc string) *RuleBuilder {
	b.rule.Description = desc
	return b
}

// WithDimension sets the allocation dimension.
func (b *RuleBuilder) WithDimension(dim Dimension) *RuleBuilder {
	b.rule.Dimension = dim
	return b
}

// WithMethod sets the allocation method.
func (b *RuleBuilder) WithMethod(method AllocationMethod) *RuleBuilder {
	b.rule.Method = method
	return b
}

// WithExpression sets the allocation expression.
func (b *RuleBuilder) WithExpression(expr string) *RuleBuilder {
	b.rule.Expression = expr
	b.rule.Method = MethodExpression
	return b
}

// AddAllocation adds a fixed allocation target.
func (b *RuleBuilder) AddAllocation(targetID string, percentage float64) *RuleBuilder {
	b.rule.Method = MethodFixed
	b.rule.Allocations = append(b.rule.Allocations, AllocationTarget{
		TargetID:   targetID,
		Percentage: percentage,
	})
	return b
}

// AddFilter adds a rule filter.
func (b *RuleBuilder) AddFilter(field, operator, value string) *RuleBuilder {
	b.rule.Filters = append(b.rule.Filters, RuleFilter{
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return b
}

// WithPriority sets the rule priority.
func (b *RuleBuilder) WithPriority(priority int) *RuleBuilder {
	b.rule.Priority = priority
	return b
}

// WithOrgID sets the organization ID.
func (b *RuleBuilder) WithOrgID(orgID string) *RuleBuilder {
	b.rule.OrgID = orgID
	return b
}

// Enabled sets whether the rule is enabled.
func (b *RuleBuilder) Enabled(enabled bool) *RuleBuilder {
	b.rule.Enabled = enabled
	return b
}

// Build returns the constructed rule after validation.
func (b *RuleBuilder) Build() (Rule, error) {
	b.rule.UpdatedAt = time.Now().UTC()

	if err := b.rule.Validate(); err != nil {
		return Rule{}, err
	}

	return b.rule, nil
}

// MustBuild returns the rule, panicking on validation error.
func (b *RuleBuilder) MustBuild() Rule {
	r, err := b.Build()
	if err != nil {
		panic(err)
	}
	return r
}

// =============================================================================
// Rule Set
// =============================================================================

// RuleSet manages a collection of allocation rules.
type RuleSet struct {
	rules   map[string]Rule
	byDim   map[Dimension][]string // dimension -> rule IDs
	ordered []string               // IDs ordered by priority
}

// NewRuleSet creates an empty rule set.
func NewRuleSet() *RuleSet {
	return &RuleSet{
		rules: make(map[string]Rule),
		byDim: make(map[Dimension][]string),
	}
}

// Add adds a rule to the set.
func (rs *RuleSet) Add(rule Rule) error {
	if err := rule.Validate(); err != nil {
		return err
	}

	rs.rules[rule.ID] = rule
	rs.byDim[rule.Dimension] = append(rs.byDim[rule.Dimension], rule.ID)
	rs.reorder()

	return nil
}

// Get retrieves a rule by ID.
func (rs *RuleSet) Get(id string) (Rule, bool) {
	r, ok := rs.rules[id]
	return r, ok
}

// Remove deletes a rule from the set.
func (rs *RuleSet) Remove(id string) {
	if rule, ok := rs.rules[id]; ok {
		delete(rs.rules, id)

		// Remove from dimension index
		dim := rule.Dimension
		ids := rs.byDim[dim]
		for i, rid := range ids {
			if rid == id {
				rs.byDim[dim] = append(ids[:i], ids[i+1:]...)
				break
			}
		}

		rs.reorder()
	}
}

// GetByDimension returns all rules for a dimension.
func (rs *RuleSet) GetByDimension(dim Dimension) []Rule {
	ids := rs.byDim[dim]
	rules := make([]Rule, 0, len(ids))
	for _, id := range ids {
		if r, ok := rs.rules[id]; ok && r.Enabled {
			rules = append(rules, r)
		}
	}
	return rules
}

// All returns all rules in priority order.
func (rs *RuleSet) All() []Rule {
	rules := make([]Rule, 0, len(rs.ordered))
	for _, id := range rs.ordered {
		if r, ok := rs.rules[id]; ok {
			rules = append(rules, r)
		}
	}
	return rules
}

// Count returns the number of rules.
func (rs *RuleSet) Count() int {
	return len(rs.rules)
}

// reorder rebuilds the priority-ordered ID list.
func (rs *RuleSet) reorder() {
	rs.ordered = make([]string, 0, len(rs.rules))
	for id := range rs.rules {
		rs.ordered = append(rs.ordered, id)
	}

	// Sort by priority (descending) then ID (for stability)
	for i := 0; i < len(rs.ordered); i++ {
		for j := i + 1; j < len(rs.ordered); j++ {
			ri := rs.rules[rs.ordered[i]]
			rj := rs.rules[rs.ordered[j]]
			if rj.Priority > ri.Priority ||
				(rj.Priority == ri.Priority && rj.ID < ri.ID) {
				rs.ordered[i], rs.ordered[j] = rs.ordered[j], rs.ordered[i]
			}
		}
	}
}
