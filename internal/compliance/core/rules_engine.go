package core

import (
	"context"
	"fmt"
	"time"
)

// ComplianceFramework represents a fully wired compliance standard
type ComplianceFramework string

const (
	FrameworkCSRD       ComplianceFramework = "CSRD"
	FrameworkSEC        ComplianceFramework = "SEC_Climate"
	FrameworkCBAM       ComplianceFramework = "CBAM"
	FrameworkCalifornia ComplianceFramework = "California_SB253"
	FrameworkIFRS       ComplianceFramework = "IFRS_S2"
)

// ComplianceInput contains data for compliance validation and reporting
type ComplianceInput struct {
	Year      int
	Framework ComplianceFramework
	Data      interface{}
	
	// Metadata for observability
	ValidationRun  time.Time
	SourceSystems  []string // AWS, Azure, GCP, SAP, etc.
	DataCompleteness float64  // 0.0 - 1.0
}

// WithData returns a copy of the input with Data set.
func (c ComplianceInput) WithData(data interface{}) ComplianceInput {
	c.Data = data
	return c
}

// ComplianceReport represents a validated compliance output
type ComplianceReport struct {
	Standard      string
	Framework     ComplianceFramework
	Content       map[string]interface{}
	ValidationResults []ValidationResult
	GeneratedAt   time.Time
	
	// Wired framework metadata
	RequiredFields    []string
	MappedDataPoints  int
	CoveragePercent   float64
}

// ValidationResult tracks compliance rule checks
type ValidationResult struct {
	Rule       string
	Passed     bool
	Message    string
	Severity   string // "error", "warning", "info"
	Framework  ComplianceFramework
}

// ComplianceMapper maps emissions to compliance reports with validation
type ComplianceMapper interface {
	BuildReport(ctx context.Context, input ComplianceInput) (ComplianceReport, error)
	ValidateInput(ctx context.Context, input ComplianceInput) ([]ValidationResult, error)
	GetRequiredFields() []string
}

// RulesEngine validates data against compliance frameworks
type RulesEngine struct {
	mappers map[ComplianceFramework]ComplianceMapper
}

// NewRulesEngine creates a fully wired compliance engine
func NewRulesEngine() *RulesEngine {
	return &RulesEngine{
		mappers: make(map[ComplianceFramework]ComplianceMapper),
	}
}

// RegisterMapper wires a compliance framework mapper into the engine
func (e *RulesEngine) RegisterMapper(framework ComplianceFramework, mapper ComplianceMapper) {
	e.mappers[framework] = mapper
}

// Validate runs compliance checks for a specific framework
func (e *RulesEngine) Validate(ctx context.Context, input ComplianceInput) ([]ValidationResult, error) {
	mapper, ok := e.mappers[input.Framework]
	if !ok {
		return nil, fmt.Errorf("compliance framework %s not wired into rules engine", input.Framework)
	}
	
	return mapper.ValidateInput(ctx, input)
}

// GenerateReport creates a compliance report with validation
func (e *RulesEngine) GenerateReport(ctx context.Context, input ComplianceInput) (ComplianceReport, error) {
	mapper, ok := e.mappers[input.Framework]
	if !ok {
		return ComplianceReport{}, fmt.Errorf("compliance framework %s not wired into rules engine", input.Framework)
	}
	
	// Run validation first
	validationResults, err := mapper.ValidateInput(ctx, input)
	if err != nil {
		return ComplianceReport{}, fmt.Errorf("validation failed for %s: %w", input.Framework, err)
	}
	
	// Generate report
	report, err := mapper.BuildReport(ctx, input)
	if err != nil {
		return ComplianceReport{}, fmt.Errorf("report generation failed for %s: %w", input.Framework, err)
	}
	
	report.ValidationResults = validationResults
	report.Framework = input.Framework
	report.GeneratedAt = time.Now()
	
	return report, nil
}
