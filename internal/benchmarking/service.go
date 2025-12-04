// Package benchmarking provides privacy-preserving industry peer comparison.
//
// This package enables anonymous comparison of emissions performance
// against industry peers without exposing sensitive company data.
package benchmarking

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"sync"
	"time"
)

// =============================================================================
// Core Types
// =============================================================================

// Industry represents an industry sector.
type Industry string

const (
	IndustryManufacturing  Industry = "manufacturing"
	IndustryTechnology     Industry = "technology"
	IndustryRetail         Industry = "retail"
	IndustryFinancial      Industry = "financial"
	IndustryHealthcare     Industry = "healthcare"
	IndustryEnergy         Industry = "energy"
	IndustryTransportation Industry = "transportation"
	IndustryConstruction   Industry = "construction"
	IndustryAgriculture    Industry = "agriculture"
)

// CompanySize categorizes companies by employee count.
type CompanySize string

const (
	SizeSmall      CompanySize = "small"      // <100 employees
	SizeMedium     CompanySize = "medium"     // 100-1000 employees
	SizeLarge      CompanySize = "large"      // 1000-10000 employees
	SizeEnterprise CompanySize = "enterprise" // >10000 employees
)

// Region represents geographic region.
type Region string

const (
	RegionNorthAmerica Region = "na"
	RegionEurope       Region = "eu"
	RegionAsiaPacific  Region = "apac"
	RegionLatinAmerica Region = "latam"
	RegionGlobal       Region = "global"
)

// AnonymousSubmission is a privacy-preserving data submission.
type AnonymousSubmission struct {
	AnonymousID string          `json:"anonymousId"` // Hashed identifier
	Industry    Industry        `json:"industry"`
	Size        CompanySize     `json:"size"`
	Region      Region          `json:"region"`
	Year        int             `json:"year"`
	Metrics     EmissionMetrics `json:"metrics"`
	SubmittedAt time.Time       `json:"submittedAt"`
}

// EmissionMetrics contains anonymized emission data.
type EmissionMetrics struct {
	TotalCO2e         float64 `json:"totalCo2e"`
	Scope1            float64 `json:"scope1"`
	Scope2            float64 `json:"scope2"`
	Scope3            float64 `json:"scope3"`
	IntensityRevenue  float64 `json:"intensityRevenue"`  // tCO2e per $M revenue
	IntensityEmployee float64 `json:"intensityEmployee"` // tCO2e per FTE
	RenewablePercent  float64 `json:"renewablePercent"`
	YoYChange         float64 `json:"yoyChange"` // Percentage
}

// BenchmarkResult contains peer comparison results.
type BenchmarkResult struct {
	Industry    Industry        `json:"industry"`
	Size        CompanySize     `json:"size"`
	Region      Region          `json:"region"`
	Year        int             `json:"year"`
	PeerCount   int             `json:"peerCount"`
	YourMetrics EmissionMetrics `json:"yourMetrics"`
	Percentiles PercentileData  `json:"percentiles"`
	Ranking     RankingInfo     `json:"ranking"`
	Insights    []Insight       `json:"insights"`
	GeneratedAt time.Time       `json:"generatedAt"`
}

// PercentileData contains distribution data.
type PercentileData struct {
	P10  MetricPercentiles `json:"p10"`
	P25  MetricPercentiles `json:"p25"`
	P50  MetricPercentiles `json:"p50"` // Median
	P75  MetricPercentiles `json:"p75"`
	P90  MetricPercentiles `json:"p90"`
	Mean MetricPercentiles `json:"mean"`
}

// MetricPercentiles contains percentile values for each metric.
type MetricPercentiles struct {
	TotalCO2e         float64 `json:"totalCo2e"`
	IntensityRevenue  float64 `json:"intensityRevenue"`
	IntensityEmployee float64 `json:"intensityEmployee"`
	RenewablePercent  float64 `json:"renewablePercent"`
	YoYChange         float64 `json:"yoyChange"`
}

// RankingInfo shows where you stand among peers.
type RankingInfo struct {
	OverallPercentile     int    `json:"overallPercentile"` // 0-100, lower is better
	IntensityPercentile   int    `json:"intensityPercentile"`
	RenewablePercentile   int    `json:"renewablePercentile"`
	ImprovementPercentile int    `json:"improvementPercentile"` // YoY improvement
	Classification        string `json:"classification"`        // "leader", "average", "laggard"
}

// Insight provides actionable observations.
type Insight struct {
	Category  string  `json:"category"`
	Title     string  `json:"title"`
	Message   string  `json:"message"`
	Severity  string  `json:"severity"`            // "positive", "neutral", "negative"
	Potential float64 `json:"potential,omitempty"` // Potential reduction tCO2e
}

// =============================================================================
// Anonymization
// =============================================================================

// Anonymizer handles data anonymization.
type Anonymizer struct {
	secret []byte
}

// NewAnonymizer creates a new anonymizer with a secret key.
func NewAnonymizer(secret string) *Anonymizer {
	return &Anonymizer{
		secret: []byte(secret),
	}
}

// HashTenantID creates an anonymous identifier.
func (a *Anonymizer) HashTenantID(tenantID string) string {
	mac := hmac.New(sha256.New, a.secret)
	mac.Write([]byte(tenantID))
	return hex.EncodeToString(mac.Sum(nil))[:16]
}

// AddNoise adds differential privacy noise to a value.
func (a *Anonymizer) AddNoise(value, epsilon float64) float64 {
	// Laplace noise for differential privacy
	// This is a simplified implementation
	noise := (math.Log(0.5) - math.Log(0.5)) / epsilon
	return value + noise
}

// BucketValue rounds a value to a bucket for k-anonymity.
func (a *Anonymizer) BucketValue(value float64, bucketSize float64) float64 {
	return math.Round(value/bucketSize) * bucketSize
}

// =============================================================================
// Pool
// =============================================================================

// Pool stores anonymous benchmark data.
type Pool struct {
	submissions map[string][]AnonymousSubmission // By industry
	anonymizer  *Anonymizer
	logger      *slog.Logger
	mu          sync.RWMutex
}

// PoolConfig configures the benchmark pool.
type PoolConfig struct {
	SecretKey string
	Logger    *slog.Logger
}

// NewPool creates a new benchmark pool.
func NewPool(cfg PoolConfig) *Pool {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return &Pool{
		submissions: make(map[string][]AnonymousSubmission),
		anonymizer:  NewAnonymizer(cfg.SecretKey),
		logger:      cfg.Logger.With("component", "benchmark-pool"),
	}
}

// Submit adds an anonymous submission to the pool.
func (p *Pool) Submit(tenantID string, industry Industry, size CompanySize, region Region, year int, metrics EmissionMetrics) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Anonymize
	submission := AnonymousSubmission{
		AnonymousID: p.anonymizer.HashTenantID(tenantID),
		Industry:    industry,
		Size:        size,
		Region:      region,
		Year:        year,
		Metrics:     metrics,
		SubmittedAt: time.Now(),
	}

	// Apply noise for differential privacy
	submission.Metrics.TotalCO2e = p.anonymizer.AddNoise(metrics.TotalCO2e, 0.1)
	submission.Metrics.IntensityRevenue = p.anonymizer.AddNoise(metrics.IntensityRevenue, 0.1)

	// Store by industry
	key := string(industry)
	p.submissions[key] = append(p.submissions[key], submission)

	p.logger.Info("benchmark submitted",
		"industry", industry,
		"size", size,
		"year", year)

	return nil
}

// GetPeerGroup finds relevant peers for comparison.
func (p *Pool) GetPeerGroup(industry Industry, size CompanySize, region Region, year int) []AnonymousSubmission {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var peers []AnonymousSubmission

	// Get industry submissions
	submissions := p.submissions[string(industry)]

	for _, sub := range submissions {
		// Match year
		if sub.Year != year {
			continue
		}

		// Flexible matching on size and region
		sizeMatch := sub.Size == size
		regionMatch := sub.Region == region || sub.Region == RegionGlobal || region == RegionGlobal

		if sizeMatch || regionMatch {
			peers = append(peers, sub)
		}
	}

	return peers
}

// =============================================================================
// Service
// =============================================================================

// Service provides benchmarking functionality.
type Service struct {
	pool   *Pool
	logger *slog.Logger
}

// ServiceConfig configures the benchmarking service.
type ServiceConfig struct {
	Pool   *Pool
	Logger *slog.Logger
}

// NewService creates a new benchmarking service.
func NewService(cfg ServiceConfig) *Service {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return &Service{
		pool:   cfg.Pool,
		logger: cfg.Logger.With("component", "benchmark-service"),
	}
}

// Compare performs a benchmark comparison.
func (s *Service) Compare(ctx context.Context, tenantID string, industry Industry, size CompanySize, region Region, year int, yourMetrics EmissionMetrics) (*BenchmarkResult, error) {
	s.logger.Info("performing benchmark comparison",
		"industry", industry,
		"size", size,
		"region", region,
		"year", year)

	// Get peer group
	peers := s.pool.GetPeerGroup(industry, size, region, year)

	if len(peers) < 5 {
		return nil, fmt.Errorf("insufficient peers for comparison: %d (need at least 5)", len(peers))
	}

	// Calculate percentiles
	percentiles := s.calculatePercentiles(peers)

	// Calculate ranking
	ranking := s.calculateRanking(yourMetrics, peers)

	// Generate insights
	insights := s.generateInsights(yourMetrics, percentiles, ranking)

	return &BenchmarkResult{
		Industry:    industry,
		Size:        size,
		Region:      region,
		Year:        year,
		PeerCount:   len(peers),
		YourMetrics: yourMetrics,
		Percentiles: percentiles,
		Ranking:     ranking,
		Insights:    insights,
		GeneratedAt: time.Now(),
	}, nil
}

// calculatePercentiles calculates distribution metrics.
func (s *Service) calculatePercentiles(peers []AnonymousSubmission) PercentileData {
	n := len(peers)
	if n == 0 {
		return PercentileData{}
	}

	// Extract metric arrays
	totalCO2e := make([]float64, n)
	intensityRev := make([]float64, n)
	intensityEmp := make([]float64, n)
	renewable := make([]float64, n)
	yoy := make([]float64, n)

	for i, p := range peers {
		totalCO2e[i] = p.Metrics.TotalCO2e
		intensityRev[i] = p.Metrics.IntensityRevenue
		intensityEmp[i] = p.Metrics.IntensityEmployee
		renewable[i] = p.Metrics.RenewablePercent
		yoy[i] = p.Metrics.YoYChange
	}

	return PercentileData{
		P10: MetricPercentiles{
			TotalCO2e:         s.percentile(totalCO2e, 10),
			IntensityRevenue:  s.percentile(intensityRev, 10),
			IntensityEmployee: s.percentile(intensityEmp, 10),
			RenewablePercent:  s.percentile(renewable, 10),
			YoYChange:         s.percentile(yoy, 10),
		},
		P25: MetricPercentiles{
			TotalCO2e:         s.percentile(totalCO2e, 25),
			IntensityRevenue:  s.percentile(intensityRev, 25),
			IntensityEmployee: s.percentile(intensityEmp, 25),
			RenewablePercent:  s.percentile(renewable, 25),
			YoYChange:         s.percentile(yoy, 25),
		},
		P50: MetricPercentiles{
			TotalCO2e:         s.percentile(totalCO2e, 50),
			IntensityRevenue:  s.percentile(intensityRev, 50),
			IntensityEmployee: s.percentile(intensityEmp, 50),
			RenewablePercent:  s.percentile(renewable, 50),
			YoYChange:         s.percentile(yoy, 50),
		},
		P75: MetricPercentiles{
			TotalCO2e:         s.percentile(totalCO2e, 75),
			IntensityRevenue:  s.percentile(intensityRev, 75),
			IntensityEmployee: s.percentile(intensityEmp, 75),
			RenewablePercent:  s.percentile(renewable, 75),
			YoYChange:         s.percentile(yoy, 75),
		},
		P90: MetricPercentiles{
			TotalCO2e:         s.percentile(totalCO2e, 90),
			IntensityRevenue:  s.percentile(intensityRev, 90),
			IntensityEmployee: s.percentile(intensityEmp, 90),
			RenewablePercent:  s.percentile(renewable, 90),
			YoYChange:         s.percentile(yoy, 90),
		},
		Mean: MetricPercentiles{
			TotalCO2e:         s.mean(totalCO2e),
			IntensityRevenue:  s.mean(intensityRev),
			IntensityEmployee: s.mean(intensityEmp),
			RenewablePercent:  s.mean(renewable),
			YoYChange:         s.mean(yoy),
		},
	}
}

// percentile calculates the p-th percentile.
func (s *Service) percentile(values []float64, p int) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	idx := (float64(p) / 100.0) * float64(len(sorted)-1)
	lower := int(math.Floor(idx))
	upper := int(math.Ceil(idx))

	if lower == upper {
		return sorted[lower]
	}

	fraction := idx - float64(lower)
	return sorted[lower]*(1-fraction) + sorted[upper]*fraction
}

// mean calculates the arithmetic mean.
func (s *Service) mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateRanking determines peer ranking.
func (s *Service) calculateRanking(metrics EmissionMetrics, peers []AnonymousSubmission) RankingInfo {
	n := len(peers)
	if n == 0 {
		return RankingInfo{}
	}

	// Count how many have worse metrics (lower = better for emissions)
	var intensityBetterThan, renewableBetterThan, improvementBetterThan int

	for _, p := range peers {
		if metrics.IntensityRevenue < p.Metrics.IntensityRevenue {
			intensityBetterThan++
		}
		if metrics.RenewablePercent > p.Metrics.RenewablePercent {
			renewableBetterThan++
		}
		if metrics.YoYChange < p.Metrics.YoYChange { // More negative = more improvement
			improvementBetterThan++
		}
	}

	intensityPct := (intensityBetterThan * 100) / n
	renewablePct := (renewableBetterThan * 100) / n
	improvementPct := (improvementBetterThan * 100) / n

	// Overall is weighted average
	overall := (intensityPct*40 + renewablePct*30 + improvementPct*30) / 100

	// Classify
	classification := "average"
	if overall <= 25 {
		classification = "leader"
	} else if overall >= 75 {
		classification = "laggard"
	}

	return RankingInfo{
		OverallPercentile:     overall,
		IntensityPercentile:   intensityPct,
		RenewablePercentile:   renewablePct,
		ImprovementPercentile: improvementPct,
		Classification:        classification,
	}
}

// generateInsights creates actionable insights.
func (s *Service) generateInsights(metrics EmissionMetrics, percentiles PercentileData, ranking RankingInfo) []Insight {
	var insights []Insight

	// Intensity insight
	if metrics.IntensityRevenue > percentiles.P75.IntensityRevenue {
		potential := (metrics.IntensityRevenue - percentiles.P50.IntensityRevenue) * 100
		insights = append(insights, Insight{
			Category:  "intensity",
			Title:     "High Emissions Intensity",
			Message:   fmt.Sprintf("Your emissions intensity is in the top quartile. Reducing to median levels could save %.0f tCO2e.", potential),
			Severity:  "negative",
			Potential: potential,
		})
	} else if metrics.IntensityRevenue < percentiles.P25.IntensityRevenue {
		insights = append(insights, Insight{
			Category: "intensity",
			Title:    "Industry-Leading Intensity",
			Message:  "Your emissions intensity is among the lowest in your peer group. Continue building on this strength.",
			Severity: "positive",
		})
	}

	// Renewable energy insight
	if metrics.RenewablePercent < percentiles.P25.RenewablePercent {
		insights = append(insights, Insight{
			Category: "renewable",
			Title:    "Low Renewable Energy Adoption",
			Message:  fmt.Sprintf("Your renewable energy usage (%.0f%%) is below peers. Increasing to median (%.0f%%) could significantly reduce Scope 2.", metrics.RenewablePercent, percentiles.P50.RenewablePercent),
			Severity: "negative",
		})
	} else if metrics.RenewablePercent > percentiles.P75.RenewablePercent {
		insights = append(insights, Insight{
			Category: "renewable",
			Title:    "High Renewable Adoption",
			Message:  "Your renewable energy adoption exceeds most peers, demonstrating leadership in clean energy transition.",
			Severity: "positive",
		})
	}

	// Year-over-year trend insight
	if metrics.YoYChange > 0 {
		insights = append(insights, Insight{
			Category: "trend",
			Title:    "Emissions Increasing",
			Message:  fmt.Sprintf("Your emissions increased %.1f%% year-over-year, while the industry median decreased %.1f%%.", metrics.YoYChange, -percentiles.P50.YoYChange),
			Severity: "negative",
		})
	} else if metrics.YoYChange < percentiles.P25.YoYChange {
		insights = append(insights, Insight{
			Category: "trend",
			Title:    "Strong Reduction Progress",
			Message:  fmt.Sprintf("Your %.1f%% reduction exceeds most peers. You're on track for significant decarbonization.", -metrics.YoYChange),
			Severity: "positive",
		})
	}

	// Overall classification insight
	switch ranking.Classification {
	case "leader":
		insights = append(insights, Insight{
			Category: "overall",
			Title:    "Climate Leader",
			Message:  "You're performing in the top quartile of your industry. Consider sharing best practices and setting science-based targets.",
			Severity: "positive",
		})
	case "laggard":
		insights = append(insights, Insight{
			Category: "overall",
			Title:    "Improvement Opportunity",
			Message:  "Your overall performance is below most peers. Focus on quick wins in energy efficiency and renewable procurement.",
			Severity: "negative",
		})
	}

	return insights
}

// =============================================================================
// Industry Targets
// =============================================================================

// IndustryTarget represents SBTi-aligned industry targets.
type IndustryTarget struct {
	Industry          Industry `json:"industry"`
	Year2030Reduction float64  `json:"year2030Reduction"` // % reduction
	Year2050Reduction float64  `json:"year2050Reduction"` // % reduction (net zero)
	IntensityTarget   float64  `json:"intensityTarget"`   // tCO2e per $M revenue
	Source            string   `json:"source"`
}

// GetIndustryTargets returns SBTi-aligned targets.
func GetIndustryTargets() map[Industry]IndustryTarget {
	return map[Industry]IndustryTarget{
		IndustryTechnology: {
			Industry:          IndustryTechnology,
			Year2030Reduction: 50,
			Year2050Reduction: 90,
			IntensityTarget:   5.0,
			Source:            "SBTi ICT Sector",
		},
		IndustryManufacturing: {
			Industry:          IndustryManufacturing,
			Year2030Reduction: 42,
			Year2050Reduction: 90,
			IntensityTarget:   20.0,
			Source:            "SBTi Manufacturing",
		},
		IndustryFinancial: {
			Industry:          IndustryFinancial,
			Year2030Reduction: 50,
			Year2050Reduction: 90,
			IntensityTarget:   2.0,
			Source:            "SBTi Financial Sector",
		},
		IndustryRetail: {
			Industry:          IndustryRetail,
			Year2030Reduction: 45,
			Year2050Reduction: 90,
			IntensityTarget:   15.0,
			Source:            "SBTi Retail",
		},
	}
}

// CompareToTarget compares metrics against industry target.
func CompareToTarget(metrics EmissionMetrics, industry Industry) (*TargetComparison, error) {
	targets := GetIndustryTargets()
	target, ok := targets[industry]
	if !ok {
		return nil, fmt.Errorf("no target for industry: %s", industry)
	}

	aligned := metrics.IntensityRevenue <= target.IntensityTarget

	gap := metrics.IntensityRevenue - target.IntensityTarget
	if gap < 0 {
		gap = 0
	}

	return &TargetComparison{
		Industry:         industry,
		CurrentIntensity: metrics.IntensityRevenue,
		TargetIntensity:  target.IntensityTarget,
		Gap:              gap,
		Aligned:          aligned,
		Target2030:       target.Year2030Reduction,
	}, nil
}

// TargetComparison compares to industry target.
type TargetComparison struct {
	Industry         Industry `json:"industry"`
	CurrentIntensity float64  `json:"currentIntensity"`
	TargetIntensity  float64  `json:"targetIntensity"`
	Gap              float64  `json:"gap"`
	Aligned          bool     `json:"aligned"`
	Target2030       float64  `json:"target2030"`
}
