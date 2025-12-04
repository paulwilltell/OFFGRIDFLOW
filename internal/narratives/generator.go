// Package narratives provides AI-powered CSRD narrative generation.
//
// This package uses LLMs to generate natural language sustainability
// report sections from structured emissions data.
package narratives

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// =============================================================================
// Report Types
// =============================================================================

// ReportSection represents a CSRD report section.
type ReportSection string

const (
	// ESRS E1 - Climate Change
	SectionE1Overview          ReportSection = "E1_overview"
	SectionE1GHGEmissions      ReportSection = "E1_ghg_emissions"
	SectionE1Targets           ReportSection = "E1_targets"
	SectionE1TransitionPlan    ReportSection = "E1_transition_plan"
	SectionE1ClimateRisks      ReportSection = "E1_climate_risks"
	SectionE1EnergyConsumption ReportSection = "E1_energy_consumption"

	// ESRS E2 - Pollution
	SectionE2Pollution ReportSection = "E2_pollution"

	// ESRS E3 - Water
	SectionE3Water ReportSection = "E3_water"

	// ESRS E4 - Biodiversity
	SectionE4Biodiversity ReportSection = "E4_biodiversity"

	// ESRS E5 - Circular Economy
	SectionE5CircularEconomy ReportSection = "E5_circular_economy"

	// Cross-cutting
	SectionExecutiveSummary ReportSection = "executive_summary"
	SectionMateriality      ReportSection = "materiality"
	SectionGovernance       ReportSection = "governance"
)

// =============================================================================
// Input Data Types
// =============================================================================

// EmissionsData contains structured emissions data for narrative generation.
type EmissionsData struct {
	TenantID        string           `json:"tenantId"`
	ReportingPeriod string           `json:"reportingPeriod"`
	Scope1          ScopeData        `json:"scope1"`
	Scope2          ScopeData        `json:"scope2"`
	Scope3          ScopeData        `json:"scope3"`
	Totals          TotalEmissions   `json:"totals"`
	Trends          []TrendData      `json:"trends,omitempty"`
	Targets         []EmissionTarget `json:"targets,omitempty"`
	CompanyProfile  *CompanyProfile  `json:"companyProfile,omitempty"`
}

// ScopeData contains scope-specific emissions.
type ScopeData struct {
	TotalCO2e   float64            `json:"totalCo2e"`
	Categories  []CategoryEmission `json:"categories"`
	YoYChange   float64            `json:"yoyChange,omitempty"` // Percentage
	Methodology string             `json:"methodology,omitempty"`
}

// CategoryEmission represents emissions by category.
type CategoryEmission struct {
	Category   string  `json:"category"`
	CO2e       float64 `json:"co2e"`
	Unit       string  `json:"unit"`
	Percentage float64 `json:"percentage"`
}

// TotalEmissions contains total emissions.
type TotalEmissions struct {
	TotalCO2e        float64 `json:"totalCo2e"`
	IntensityRevenue float64 `json:"intensityRevenue,omitempty"` // tCO2e per $M revenue
	IntensityFTE     float64 `json:"intensityFte,omitempty"`     // tCO2e per FTE
	BaselineYear     string  `json:"baselineYear,omitempty"`
	BaselineCO2e     float64 `json:"baselineCo2e,omitempty"`
}

// TrendData contains historical trend information.
type TrendData struct {
	Period string  `json:"period"`
	CO2e   float64 `json:"co2e"`
	Scope1 float64 `json:"scope1,omitempty"`
	Scope2 float64 `json:"scope2,omitempty"`
	Scope3 float64 `json:"scope3,omitempty"`
}

// EmissionTarget represents a reduction target.
type EmissionTarget struct {
	Name          string  `json:"name"`
	TargetYear    int     `json:"targetYear"`
	BaselineYear  int     `json:"baselineYear"`
	BaselineCO2e  float64 `json:"baselineCo2e"`
	TargetCO2e    float64 `json:"targetCo2e"`
	Reduction     float64 `json:"reduction"` // Percentage
	Scope         string  `json:"scope"`     // "1", "2", "3", "all"
	SBTiValidated bool    `json:"sbtiValidated"`
	Progress      float64 `json:"progress"` // Percentage complete
}

// CompanyProfile provides company context.
type CompanyProfile struct {
	Name      string   `json:"name"`
	Industry  string   `json:"industry"`
	Employees int      `json:"employees"`
	Revenue   float64  `json:"revenue"`
	Locations int      `json:"locations"`
	Countries []string `json:"countries,omitempty"`
}

// =============================================================================
// Output Types
// =============================================================================

// GeneratedNarrative is the LLM-generated text.
type GeneratedNarrative struct {
	Section     ReportSection `json:"section"`
	Title       string        `json:"title"`
	Content     string        `json:"content"`
	DataSources []string      `json:"dataSources,omitempty"`
	GeneratedAt time.Time     `json:"generatedAt"`
	ModelUsed   string        `json:"modelUsed,omitempty"`
	Confidence  float64       `json:"confidence,omitempty"`
	WordCount   int           `json:"wordCount"`
}

// =============================================================================
// Generator
// =============================================================================

// Generator creates AI-powered narratives.
type Generator struct {
	llmClient LLMClient
	templates map[ReportSection]string
	logger    *slog.Logger
}

// LLMClient defines the interface to the LLM provider.
type LLMClient interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

// GeneratorConfig configures the narrative generator.
type GeneratorConfig struct {
	LLMClient LLMClient
	Logger    *slog.Logger
}

// NewGenerator creates a new narrative generator.
func NewGenerator(cfg GeneratorConfig) *Generator {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	g := &Generator{
		llmClient: cfg.LLMClient,
		templates: make(map[ReportSection]string),
		logger:    cfg.Logger.With("component", "narrative-generator"),
	}

	g.initTemplates()
	return g
}

// initTemplates sets up section-specific prompt templates.
func (g *Generator) initTemplates() {
	g.templates[SectionExecutiveSummary] = `Generate an executive summary for a CSRD sustainability report.

Company: {{.CompanyName}}
Industry: {{.Industry}}
Reporting Period: {{.ReportingPeriod}}

Key Metrics:
- Total GHG Emissions: {{.TotalCO2e}} tCO2e
- Scope 1: {{.Scope1CO2e}} tCO2e
- Scope 2: {{.Scope2CO2e}} tCO2e  
- Scope 3: {{.Scope3CO2e}} tCO2e
- Year-over-Year Change: {{.YoYChange}}%
- Emissions Intensity: {{.IntensityRevenue}} tCO2e per €M revenue

Targets:
{{.TargetsSummary}}

Write a professional 200-300 word executive summary covering:
1. Overall emissions performance
2. Key achievements and challenges
3. Progress toward targets
4. Strategic priorities for decarbonization

Use formal business language appropriate for annual reports and investor communications.`

	g.templates[SectionE1GHGEmissions] = `Generate ESRS E1 GHG emissions disclosure text for a CSRD report.

Company: {{.CompanyName}}
Reporting Period: {{.ReportingPeriod}}

SCOPE 1 EMISSIONS (Direct):
Total: {{.Scope1CO2e}} tCO2e
Categories:
{{.Scope1Categories}}

SCOPE 2 EMISSIONS (Indirect - Energy):
Location-based: {{.Scope2LocationCO2e}} tCO2e
Market-based: {{.Scope2MarketCO2e}} tCO2e
Categories:
{{.Scope2Categories}}

SCOPE 3 EMISSIONS (Value Chain):
Total: {{.Scope3CO2e}} tCO2e
Categories:
{{.Scope3Categories}}

Methodology: {{.Methodology}}

Generate professional CSRD-compliant disclosure text (300-500 words) covering:
1. Gross Scope 1, 2, and 3 GHG emissions
2. Significant changes from prior periods
3. Emissions intensity metrics
4. Methodology and data quality
5. GHG Protocol alignment

Follow ESRS E1-6 disclosure requirements.`

	g.templates[SectionE1Targets] = `Generate ESRS E1 climate targets disclosure for a CSRD report.

Company: {{.CompanyName}}
Baseline Year: {{.BaselineYear}}
Baseline Emissions: {{.BaselineCO2e}} tCO2e

REDUCTION TARGETS:
{{.TargetsDetail}}

Current Progress: {{.OverallProgress}}% toward targets

Generate professional disclosure text (200-400 words) covering:
1. GHG reduction targets with timelines
2. Science-based validation status (SBTi)
3. Progress against baseline
4. Key decarbonization initiatives
5. Scope coverage and methodology

Follow ESRS E1-4 target disclosure requirements.`

	g.templates[SectionE1TransitionPlan] = `Generate ESRS E1 climate transition plan summary for a CSRD report.

Company: {{.CompanyName}}
Industry: {{.Industry}}

Current Emissions Profile:
- Scope 1: {{.Scope1CO2e}} tCO2e
- Scope 2: {{.Scope2CO2e}} tCO2e
- Scope 3: {{.Scope3CO2e}} tCO2e

Reduction Targets:
{{.TargetsSummary}}

Generate a climate transition plan summary (400-600 words) covering:
1. Paris Agreement alignment strategy
2. Decarbonization levers by scope
3. Key milestones and checkpoints
4. Investment requirements and funding
5. Technology and operational changes
6. Supply chain engagement approach
7. Carbon offsetting strategy (if applicable)
8. Governance and oversight

Follow ESRS E1-1 transition plan requirements.`

	g.templates[SectionE1ClimateRisks] = `Generate ESRS E1 climate risks and opportunities disclosure.

Company: {{.CompanyName}}
Industry: {{.Industry}}
Geographic Presence: {{.Countries}}

Generate professional disclosure text (300-500 words) covering:
1. Physical climate risks (acute and chronic)
2. Transition risks (policy, technology, market, reputation)
3. Climate-related opportunities
4. Financial implications
5. Scenario analysis approach

Follow TCFD recommendations and ESRS E1-9 requirements.`

	g.templates[SectionMateriality] = `Generate materiality assessment disclosure for a CSRD report.

Company: {{.CompanyName}}
Industry: {{.Industry}}

Environmental Metrics:
- Total GHG Emissions: {{.TotalCO2e}} tCO2e
- Energy Consumption: {{.EnergyConsumption}} MWh
- Emissions Intensity: {{.IntensityRevenue}} tCO2e/€M

Generate double materiality assessment text (300-500 words) covering:
1. Material sustainability matters identification
2. Impact materiality assessment
3. Financial materiality assessment  
4. Stakeholder engagement in process
5. How climate change affects the company (outside-in)
6. How the company affects climate (inside-out)

Follow ESRS 1 and ESRS 2 materiality requirements.`
}

// Generate creates a narrative for a specific section.
func (g *Generator) Generate(ctx context.Context, section ReportSection, data EmissionsData) (*GeneratedNarrative, error) {
	template, ok := g.templates[section]
	if !ok {
		return nil, fmt.Errorf("no template for section: %s", section)
	}

	// Build prompt from template and data
	prompt := g.buildPrompt(template, data)

	g.logger.Info("generating narrative",
		"section", section,
		"tenant", data.TenantID)

	// Call LLM
	content, err := g.llmClient.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	return &GeneratedNarrative{
		Section:     section,
		Title:       g.getSectionTitle(section),
		Content:     content,
		GeneratedAt: time.Now(),
		WordCount:   len(strings.Fields(content)),
	}, nil
}

// buildPrompt populates the template with data.
func (g *Generator) buildPrompt(template string, data EmissionsData) string {
	replacements := map[string]string{
		"{{.CompanyName}}":      data.CompanyProfile.Name,
		"{{.Industry}}":         data.CompanyProfile.Industry,
		"{{.ReportingPeriod}}":  data.ReportingPeriod,
		"{{.TotalCO2e}}":        fmt.Sprintf("%.2f", data.Totals.TotalCO2e),
		"{{.Scope1CO2e}}":       fmt.Sprintf("%.2f", data.Scope1.TotalCO2e),
		"{{.Scope2CO2e}}":       fmt.Sprintf("%.2f", data.Scope2.TotalCO2e),
		"{{.Scope3CO2e}}":       fmt.Sprintf("%.2f", data.Scope3.TotalCO2e),
		"{{.YoYChange}}":        fmt.Sprintf("%.1f", g.calcOverallYoYChange(data)),
		"{{.IntensityRevenue}}": fmt.Sprintf("%.2f", data.Totals.IntensityRevenue),
		"{{.Scope1Categories}}": g.formatCategories(data.Scope1.Categories),
		"{{.Scope2Categories}}": g.formatCategories(data.Scope2.Categories),
		"{{.Scope3Categories}}": g.formatCategories(data.Scope3.Categories),
		"{{.TargetsSummary}}":   g.formatTargetsSummary(data.Targets),
		"{{.TargetsDetail}}":    g.formatTargetsDetail(data.Targets),
		"{{.Methodology}}":      data.Scope1.Methodology,
	}

	if data.CompanyProfile != nil {
		replacements["{{.Countries}}"] = strings.Join(data.CompanyProfile.Countries, ", ")
	}

	if len(data.Targets) > 0 {
		replacements["{{.BaselineYear}}"] = fmt.Sprintf("%d", data.Targets[0].BaselineYear)
		replacements["{{.BaselineCO2e}}"] = fmt.Sprintf("%.2f", data.Targets[0].BaselineCO2e)
		replacements["{{.OverallProgress}}"] = fmt.Sprintf("%.1f", g.calcOverallProgress(data.Targets))
	}

	result := template
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// Helper methods for formatting.
func (g *Generator) formatCategories(cats []CategoryEmission) string {
	var lines []string
	for _, c := range cats {
		lines = append(lines, fmt.Sprintf("- %s: %.2f tCO2e (%.1f%%)",
			c.Category, c.CO2e, c.Percentage))
	}
	return strings.Join(lines, "\n")
}

func (g *Generator) formatTargetsSummary(targets []EmissionTarget) string {
	var lines []string
	for _, t := range targets {
		status := "In Progress"
		if t.SBTiValidated {
			status = "SBTi Validated"
		}
		lines = append(lines, fmt.Sprintf("- %s: %.0f%% reduction by %d (%s)",
			t.Name, t.Reduction, t.TargetYear, status))
	}
	return strings.Join(lines, "\n")
}

func (g *Generator) formatTargetsDetail(targets []EmissionTarget) string {
	var lines []string
	for _, t := range targets {
		lines = append(lines, fmt.Sprintf(`
Target: %s
- Scope: %s
- Baseline (%d): %.2f tCO2e
- Target (%d): %.2f tCO2e  
- Reduction: %.0f%%
- Progress: %.1f%% achieved
- SBTi Validated: %v`,
			t.Name, t.Scope, t.BaselineYear, t.BaselineCO2e,
			t.TargetYear, t.TargetCO2e, t.Reduction, t.Progress, t.SBTiValidated))
	}
	return strings.Join(lines, "\n")
}

func (g *Generator) calcOverallYoYChange(data EmissionsData) float64 {
	if len(data.Trends) < 2 {
		return 0
	}
	prev := data.Trends[len(data.Trends)-2].CO2e
	curr := data.Trends[len(data.Trends)-1].CO2e
	if prev == 0 {
		return 0
	}
	return ((curr - prev) / prev) * 100
}

func (g *Generator) calcOverallProgress(targets []EmissionTarget) float64 {
	if len(targets) == 0 {
		return 0
	}
	var total float64
	for _, t := range targets {
		total += t.Progress
	}
	return total / float64(len(targets))
}

func (g *Generator) getSectionTitle(section ReportSection) string {
	titles := map[ReportSection]string{
		SectionExecutiveSummary:    "Executive Summary",
		SectionE1Overview:          "E1 Climate Change Overview",
		SectionE1GHGEmissions:      "E1-6 Gross Scopes 1, 2, 3 and Total GHG Emissions",
		SectionE1Targets:           "E1-4 GHG Emission Reduction Targets",
		SectionE1TransitionPlan:    "E1-1 Transition Plan for Climate Change Mitigation",
		SectionE1ClimateRisks:      "E1-9 Anticipated Financial Effects from Climate Risks",
		SectionE1EnergyConsumption: "E1-5 Energy Consumption and Mix",
		SectionMateriality:         "Double Materiality Assessment",
		SectionGovernance:          "Climate Governance Structure",
	}
	if title, ok := titles[section]; ok {
		return title
	}
	return string(section)
}

// GenerateFullReport generates all sections.
func (g *Generator) GenerateFullReport(ctx context.Context, data EmissionsData) ([]GeneratedNarrative, error) {
	sections := []ReportSection{
		SectionExecutiveSummary,
		SectionMateriality,
		SectionE1GHGEmissions,
		SectionE1Targets,
		SectionE1TransitionPlan,
		SectionE1ClimateRisks,
	}

	var narratives []GeneratedNarrative
	for _, section := range sections {
		narrative, err := g.Generate(ctx, section, data)
		if err != nil {
			g.logger.Warn("failed to generate section",
				"section", section,
				"error", err)
			continue
		}
		narratives = append(narratives, *narrative)
	}

	return narratives, nil
}

// =============================================================================
// Mock LLM Client for Testing
// =============================================================================

// MockLLMClient provides a test implementation.
type MockLLMClient struct{}

// Complete returns mock generated text.
func (m *MockLLMClient) Complete(ctx context.Context, prompt string) (string, error) {
	// Return realistic mock content for testing
	if strings.Contains(prompt, "executive summary") {
		return `In the reporting period, our organization demonstrated continued commitment to climate action while facing operational challenges. Total greenhouse gas emissions reached 15,432 tCO2e, representing a 5% reduction from the prior year baseline.

Scope 1 direct emissions from owned operations accounted for 2,100 tCO2e, primarily from fleet vehicles and on-site combustion. We achieved a 12% reduction through fleet electrification and improved energy efficiency.

Scope 2 emissions from purchased electricity totaled 4,200 tCO2e (location-based). Our renewable energy procurement strategy reduced market-based Scope 2 emissions by 35% through the purchase of renewable energy certificates.

Scope 3 value chain emissions remain our largest category at 9,132 tCO2e, with purchased goods and services and business travel being the primary contributors. We have initiated supplier engagement programs to address these emissions.

Progress toward our 2030 target of 50% reduction remains on track, with 23% reduction achieved to date. Key priorities for the coming year include expanding renewable energy procurement, accelerating fleet electrification, and deepening supply chain engagement.`, nil
	}

	return "Generated CSRD-compliant disclosure text based on provided emissions data.", nil
}

// =============================================================================
// Report Assembly
// =============================================================================

// ReportAssembler combines narratives into a complete report.
type ReportAssembler struct {
	generator *Generator
	logger    *slog.Logger
}

// AssembledReport is the complete CSRD report.
type AssembledReport struct {
	TenantID        string               `json:"tenantId"`
	Title           string               `json:"title"`
	ReportingPeriod string               `json:"reportingPeriod"`
	Sections        []GeneratedNarrative `json:"sections"`
	GeneratedAt     time.Time            `json:"generatedAt"`
	Version         string               `json:"version"`
	TotalWordCount  int                  `json:"totalWordCount"`
}

// NewReportAssembler creates a new assembler.
func NewReportAssembler(generator *Generator, logger *slog.Logger) *ReportAssembler {
	return &ReportAssembler{
		generator: generator,
		logger:    logger.With("component", "report-assembler"),
	}
}

// Assemble creates a complete CSRD report.
func (ra *ReportAssembler) Assemble(ctx context.Context, data EmissionsData) (*AssembledReport, error) {
	ra.logger.Info("assembling CSRD report",
		"tenant", data.TenantID,
		"period", data.ReportingPeriod)

	narratives, err := ra.generator.GenerateFullReport(ctx, data)
	if err != nil {
		return nil, err
	}

	var totalWords int
	for _, n := range narratives {
		totalWords += n.WordCount
	}

	companyName := "Company"
	if data.CompanyProfile != nil {
		companyName = data.CompanyProfile.Name
	}

	return &AssembledReport{
		TenantID:        data.TenantID,
		Title:           fmt.Sprintf("%s CSRD Sustainability Report %s", companyName, data.ReportingPeriod),
		ReportingPeriod: data.ReportingPeriod,
		Sections:        narratives,
		GeneratedAt:     time.Now(),
		Version:         "1.0",
		TotalWordCount:  totalWords,
	}, nil
}

// ExportMarkdown exports the report as Markdown.
func (ra *ReportAssembler) ExportMarkdown(report *AssembledReport) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", report.Title))
	sb.WriteString(fmt.Sprintf("**Reporting Period:** %s\n\n", report.ReportingPeriod))
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", report.GeneratedAt.Format("January 2, 2006")))
	sb.WriteString("---\n\n")

	for _, section := range report.Sections {
		sb.WriteString(fmt.Sprintf("## %s\n\n", section.Title))
		sb.WriteString(section.Content)
		sb.WriteString("\n\n---\n\n")
	}

	return sb.String()
}

// ExportJSON exports the report as JSON.
func (ra *ReportAssembler) ExportJSON(report *AssembledReport) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}
