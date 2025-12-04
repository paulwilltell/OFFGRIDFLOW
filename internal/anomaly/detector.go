// Package anomaly provides AI-powered anomaly detection for emissions data.
//
// This package uses statistical methods and machine learning patterns to identify
// unusual emissions patterns, data quality issues, and reduction opportunities.
package anomaly

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"
)

// =============================================================================
// Anomaly Types
// =============================================================================

// AnomalyType categorizes detected anomalies.
type AnomalyType string

const (
	// AnomalyTypeSpike indicates a sudden increase in emissions.
	AnomalyTypeSpike AnomalyType = "spike"

	// AnomalyTypeDrop indicates an unexpected decrease in emissions.
	AnomalyTypeDrop AnomalyType = "drop"

	// AnomalyTypeTrend indicates a concerning long-term trend.
	AnomalyTypeTrend AnomalyType = "trend"

	// AnomalyTypeSeasonality indicates deviation from expected seasonal pattern.
	AnomalyTypeSeasonality AnomalyType = "seasonality"

	// AnomalyTypeDataQuality indicates potential data quality issues.
	AnomalyTypeDataQuality AnomalyType = "data_quality"

	// AnomalyTypeThreshold indicates a threshold breach.
	AnomalyTypeThreshold AnomalyType = "threshold"
)

// Severity indicates how critical an anomaly is.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

// Anomaly represents a detected anomaly in emissions data.
type Anomaly struct {
	ID          string                 `json:"id"`
	Type        AnomalyType            `json:"type"`
	Severity    Severity               `json:"severity"`
	TenantID    string                 `json:"tenantId"`
	Scope       int                    `json:"scope,omitempty"`
	Category    string                 `json:"category,omitempty"`
	Source      string                 `json:"source,omitempty"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Value       float64                `json:"value"`
	Expected    float64                `json:"expected"`
	Deviation   float64                `json:"deviation"` // Standard deviations from mean
	DetectedAt  time.Time              `json:"detectedAt"`
	PeriodStart time.Time              `json:"periodStart"`
	PeriodEnd   time.Time              `json:"periodEnd"`
	Confidence  float64                `json:"confidence"` // 0-1
	Suggestion  string                 `json:"suggestion,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// =============================================================================
// Data Point
// =============================================================================

// DataPoint represents a single emissions measurement.
type DataPoint struct {
	Timestamp time.Time
	Value     float64
	Scope     int
	Category  string
	Source    string
	Region    string
}

// TimeSeries represents a sequence of data points.
type TimeSeries struct {
	TenantID string
	Points   []DataPoint
}

// =============================================================================
// Detection Configuration
// =============================================================================

// Config holds anomaly detection configuration.
type Config struct {
	// MinDataPoints is the minimum history required for detection.
	MinDataPoints int

	// SpikeThreshold is the z-score threshold for spike detection.
	SpikeThreshold float64

	// DropThreshold is the z-score threshold for drop detection.
	DropThreshold float64

	// TrendWindowDays is the window for trend analysis.
	TrendWindowDays int

	// SeasonalityPeriod is the expected seasonal period in days.
	SeasonalityPeriod int

	// DataQualityThreshold is the threshold for data quality issues.
	DataQualityThreshold float64

	// Thresholds are custom thresholds per scope/category.
	Thresholds map[string]float64

	// Logger for detection operations.
	Logger *slog.Logger
}

// DefaultConfig returns sensible default configuration.
func DefaultConfig() Config {
	return Config{
		MinDataPoints:        30,
		SpikeThreshold:       3.0, // 3 standard deviations
		DropThreshold:        -2.5,
		TrendWindowDays:      90,
		SeasonalityPeriod:    365,
		DataQualityThreshold: 0.1, // 10% zero/missing values
		Thresholds:           make(map[string]float64),
		Logger:               slog.Default(),
	}
}

// =============================================================================
// Detector
// =============================================================================

// Detector performs anomaly detection on emissions data.
type Detector struct {
	config Config
	logger *slog.Logger
}

// NewDetector creates a new anomaly detector.
func NewDetector(config Config) *Detector {
	if config.Logger == nil {
		config.Logger = slog.Default()
	}
	return &Detector{
		config: config,
		logger: config.Logger.With("component", "anomaly-detector"),
	}
}

// Detect analyzes a time series and returns detected anomalies.
func (d *Detector) Detect(ctx context.Context, series TimeSeries) ([]Anomaly, error) {
	if len(series.Points) < d.config.MinDataPoints {
		return nil, fmt.Errorf("insufficient data: need %d points, have %d",
			d.config.MinDataPoints, len(series.Points))
	}

	var anomalies []Anomaly

	// Run all detection methods
	spikes := d.detectSpikes(series)
	anomalies = append(anomalies, spikes...)

	drops := d.detectDrops(series)
	anomalies = append(anomalies, drops...)

	trends := d.detectTrends(series)
	anomalies = append(anomalies, trends...)

	quality := d.detectDataQualityIssues(series)
	anomalies = append(anomalies, quality...)

	thresholds := d.detectThresholdBreaches(series)
	anomalies = append(anomalies, thresholds...)

	// Sort by severity and time
	sort.Slice(anomalies, func(i, j int) bool {
		if anomalies[i].Severity != anomalies[j].Severity {
			return severityOrder(anomalies[i].Severity) < severityOrder(anomalies[j].Severity)
		}
		return anomalies[i].DetectedAt.After(anomalies[j].DetectedAt)
	})

	return anomalies, nil
}

// DetectRealtime performs lightweight detection on a single new data point.
func (d *Detector) DetectRealtime(ctx context.Context, tenantID string, point DataPoint, history []DataPoint) []Anomaly {
	if len(history) < 7 {
		return nil
	}

	var anomalies []Anomaly

	// Calculate statistics from history
	values := make([]float64, len(history))
	for i, p := range history {
		values[i] = p.Value
	}

	mean := calculateMean(values)
	stdDev := calculateStdDev(values, mean)

	if stdDev == 0 {
		return nil
	}

	// Check for spike
	zScore := (point.Value - mean) / stdDev
	if zScore > d.config.SpikeThreshold {
		anomalies = append(anomalies, Anomaly{
			ID:          fmt.Sprintf("spike-%d", time.Now().UnixNano()),
			Type:        AnomalyTypeSpike,
			Severity:    d.calculateSeverity(zScore),
			TenantID:    tenantID,
			Scope:       point.Scope,
			Category:    point.Category,
			Source:      point.Source,
			Title:       fmt.Sprintf("Emission Spike Detected: %.1fx above normal", zScore),
			Description: fmt.Sprintf("Current value %.2f is %.1f standard deviations above the mean of %.2f", point.Value, zScore, mean),
			Value:       point.Value,
			Expected:    mean,
			Deviation:   zScore,
			DetectedAt:  time.Now(),
			PeriodStart: point.Timestamp,
			PeriodEnd:   point.Timestamp,
			Confidence:  d.calculateConfidence(zScore, len(history)),
			Suggestion:  d.suggestAction(AnomalyTypeSpike, zScore, point),
		})
	}

	// Check for drop
	if zScore < d.config.DropThreshold {
		anomalies = append(anomalies, Anomaly{
			ID:          fmt.Sprintf("drop-%d", time.Now().UnixNano()),
			Type:        AnomalyTypeDrop,
			Severity:    SeverityMedium,
			TenantID:    tenantID,
			Scope:       point.Scope,
			Category:    point.Category,
			Source:      point.Source,
			Title:       fmt.Sprintf("Unusual Drop Detected: %.1fx below normal", -zScore),
			Description: fmt.Sprintf("Current value %.2f is %.1f standard deviations below the mean of %.2f", point.Value, -zScore, mean),
			Value:       point.Value,
			Expected:    mean,
			Deviation:   zScore,
			DetectedAt:  time.Now(),
			PeriodStart: point.Timestamp,
			PeriodEnd:   point.Timestamp,
			Confidence:  d.calculateConfidence(-zScore, len(history)),
			Suggestion:  d.suggestAction(AnomalyTypeDrop, zScore, point),
		})
	}

	return anomalies
}

// =============================================================================
// Detection Methods
// =============================================================================

func (d *Detector) detectSpikes(series TimeSeries) []Anomaly {
	var anomalies []Anomaly

	values := extractValues(series.Points)
	mean := calculateMean(values)
	stdDev := calculateStdDev(values, mean)

	if stdDev == 0 {
		return nil
	}

	for _, point := range series.Points {
		zScore := (point.Value - mean) / stdDev
		if zScore > d.config.SpikeThreshold {
			anomalies = append(anomalies, Anomaly{
				ID:          fmt.Sprintf("spike-%s-%d", series.TenantID, point.Timestamp.UnixNano()),
				Type:        AnomalyTypeSpike,
				Severity:    d.calculateSeverity(zScore),
				TenantID:    series.TenantID,
				Scope:       point.Scope,
				Category:    point.Category,
				Source:      point.Source,
				Title:       "Emission Spike Detected",
				Description: fmt.Sprintf("Value of %.2f kg CO2e is %.1f standard deviations above normal", point.Value, zScore),
				Value:       point.Value,
				Expected:    mean,
				Deviation:   zScore,
				DetectedAt:  time.Now(),
				PeriodStart: point.Timestamp,
				PeriodEnd:   point.Timestamp,
				Confidence:  d.calculateConfidence(zScore, len(series.Points)),
				Suggestion:  d.suggestAction(AnomalyTypeSpike, zScore, point),
			})
		}
	}

	return anomalies
}

func (d *Detector) detectDrops(series TimeSeries) []Anomaly {
	var anomalies []Anomaly

	values := extractValues(series.Points)
	mean := calculateMean(values)
	stdDev := calculateStdDev(values, mean)

	if stdDev == 0 {
		return nil
	}

	for _, point := range series.Points {
		zScore := (point.Value - mean) / stdDev
		if zScore < d.config.DropThreshold {
			anomalies = append(anomalies, Anomaly{
				ID:          fmt.Sprintf("drop-%s-%d", series.TenantID, point.Timestamp.UnixNano()),
				Type:        AnomalyTypeDrop,
				Severity:    SeverityMedium, // Drops are often positive (efficiency gains)
				TenantID:    series.TenantID,
				Scope:       point.Scope,
				Category:    point.Category,
				Title:       "Unusual Emission Drop",
				Description: fmt.Sprintf("Value of %.2f kg CO2e is %.1f standard deviations below normal", point.Value, -zScore),
				Value:       point.Value,
				Expected:    mean,
				Deviation:   zScore,
				DetectedAt:  time.Now(),
				PeriodStart: point.Timestamp,
				PeriodEnd:   point.Timestamp,
				Confidence:  d.calculateConfidence(-zScore, len(series.Points)),
				Suggestion:  d.suggestAction(AnomalyTypeDrop, zScore, point),
			})
		}
	}

	return anomalies
}

func (d *Detector) detectTrends(series TimeSeries) []Anomaly {
	var anomalies []Anomaly

	if len(series.Points) < d.config.TrendWindowDays {
		return nil
	}

	// Simple linear regression for trend detection
	slope, rSquared := calculateLinearTrend(series.Points)

	// Significant upward trend
	if slope > 0 && rSquared > 0.7 {
		values := extractValues(series.Points)
		mean := calculateMean(values)
		percentIncrease := (slope * float64(d.config.TrendWindowDays) / mean) * 100

		if percentIncrease > 10 { // More than 10% increase over window
			anomalies = append(anomalies, Anomaly{
				ID:          fmt.Sprintf("trend-up-%s-%d", series.TenantID, time.Now().UnixNano()),
				Type:        AnomalyTypeTrend,
				Severity:    d.trendSeverity(percentIncrease),
				TenantID:    series.TenantID,
				Title:       "Rising Emissions Trend",
				Description: fmt.Sprintf("Emissions are increasing at %.1f%% over %d days (R²=%.2f)", percentIncrease, d.config.TrendWindowDays, rSquared),
				Value:       slope,
				Expected:    0, // Flat trend expected
				Deviation:   percentIncrease,
				DetectedAt:  time.Now(),
				PeriodStart: series.Points[0].Timestamp,
				PeriodEnd:   series.Points[len(series.Points)-1].Timestamp,
				Confidence:  rSquared,
				Suggestion:  "Investigate operational changes driving increased emissions. Consider efficiency measures.",
			})
		}
	}

	return anomalies
}

func (d *Detector) detectDataQualityIssues(series TimeSeries) []Anomaly {
	var anomalies []Anomaly

	// Check for gaps
	gaps := d.findDataGaps(series.Points)
	if len(gaps) > 0 {
		anomalies = append(anomalies, Anomaly{
			ID:          fmt.Sprintf("gaps-%s-%d", series.TenantID, time.Now().UnixNano()),
			Type:        AnomalyTypeDataQuality,
			Severity:    SeverityMedium,
			TenantID:    series.TenantID,
			Title:       "Data Gaps Detected",
			Description: fmt.Sprintf("Found %d gaps in emissions data", len(gaps)),
			DetectedAt:  time.Now(),
			Suggestion:  "Review data collection processes to ensure continuous monitoring.",
			Metadata:    map[string]interface{}{"gaps": gaps},
		})
	}

	// Check for zeros
	zeroCount := 0
	for _, p := range series.Points {
		if p.Value == 0 {
			zeroCount++
		}
	}
	zeroRatio := float64(zeroCount) / float64(len(series.Points))
	if zeroRatio > d.config.DataQualityThreshold {
		anomalies = append(anomalies, Anomaly{
			ID:          fmt.Sprintf("zeros-%s-%d", series.TenantID, time.Now().UnixNano()),
			Type:        AnomalyTypeDataQuality,
			Severity:    SeverityHigh,
			TenantID:    series.TenantID,
			Title:       "High Zero Value Rate",
			Description: fmt.Sprintf("%.1f%% of data points are zero", zeroRatio*100),
			Value:       zeroRatio,
			DetectedAt:  time.Now(),
			Suggestion:  "Verify data sources are reporting correctly. Zero values may indicate collection failures.",
		})
	}

	// Check for duplicates
	duplicates := d.findDuplicates(series.Points)
	if len(duplicates) > 0 {
		anomalies = append(anomalies, Anomaly{
			ID:          fmt.Sprintf("dupes-%s-%d", series.TenantID, time.Now().UnixNano()),
			Type:        AnomalyTypeDataQuality,
			Severity:    SeverityMedium,
			TenantID:    series.TenantID,
			Title:       "Duplicate Data Points",
			Description: fmt.Sprintf("Found %d duplicate timestamps", len(duplicates)),
			DetectedAt:  time.Now(),
			Suggestion:  "Review ingestion pipelines for double-counting issues.",
		})
	}

	return anomalies
}

func (d *Detector) detectThresholdBreaches(series TimeSeries) []Anomaly {
	var anomalies []Anomaly

	for _, point := range series.Points {
		key := fmt.Sprintf("scope%d:%s", point.Scope, point.Category)
		if threshold, ok := d.config.Thresholds[key]; ok {
			if point.Value > threshold {
				anomalies = append(anomalies, Anomaly{
					ID:          fmt.Sprintf("threshold-%s-%d", series.TenantID, point.Timestamp.UnixNano()),
					Type:        AnomalyTypeThreshold,
					Severity:    SeverityHigh,
					TenantID:    series.TenantID,
					Scope:       point.Scope,
					Category:    point.Category,
					Title:       "Threshold Exceeded",
					Description: fmt.Sprintf("Value %.2f exceeds threshold of %.2f", point.Value, threshold),
					Value:       point.Value,
					Expected:    threshold,
					Deviation:   (point.Value - threshold) / threshold * 100,
					DetectedAt:  time.Now(),
					PeriodStart: point.Timestamp,
					PeriodEnd:   point.Timestamp,
					Confidence:  1.0,
					Suggestion:  "Immediate action required to reduce emissions below threshold.",
				})
			}
		}
	}

	return anomalies
}

// =============================================================================
// Helper Functions
// =============================================================================

func (d *Detector) calculateSeverity(zScore float64) Severity {
	absZ := math.Abs(zScore)
	switch {
	case absZ >= 5:
		return SeverityCritical
	case absZ >= 4:
		return SeverityHigh
	case absZ >= 3:
		return SeverityMedium
	default:
		return SeverityLow
	}
}

func (d *Detector) trendSeverity(percentChange float64) Severity {
	switch {
	case percentChange >= 50:
		return SeverityCritical
	case percentChange >= 25:
		return SeverityHigh
	case percentChange >= 10:
		return SeverityMedium
	default:
		return SeverityLow
	}
}

func (d *Detector) calculateConfidence(zScore float64, sampleSize int) float64 {
	// Higher z-score and larger sample = higher confidence
	zConf := math.Min(math.Abs(zScore)/5.0, 1.0)
	sizeConf := math.Min(float64(sampleSize)/100.0, 1.0)
	return (zConf + sizeConf) / 2
}

func (d *Detector) suggestAction(anomalyType AnomalyType, deviation float64, point DataPoint) string {
	severity := "moderate"
	if math.Abs(deviation) > 4.0 {
		severity = "severe"
	} else if math.Abs(deviation) > 2.0 {
		severity = "significant"
	}

	switch anomalyType {
	case AnomalyTypeSpike:
		if point.Source != "" {
			return fmt.Sprintf("Investigate %s data source for unusual activity (%s deviation: %.1f σ). Check for data quality issues or operational changes.", point.Source, severity, deviation)
		}
		return fmt.Sprintf("Investigate cause of emission spike (%s deviation: %.1f σ). May indicate operational change, data quality issue, or reporting error.", severity, deviation)
	case AnomalyTypeDrop:
		return fmt.Sprintf("Verify data completeness (%s deviation: %.1f σ). If confirmed, document efficiency gains for reporting.", severity, deviation)
	default:
		return fmt.Sprintf("Review data and investigate root cause (deviation: %.1f σ).", deviation)
	}
}

func (d *Detector) findDataGaps(points []DataPoint) []time.Duration {
	var gaps []time.Duration
	if len(points) < 2 {
		return gaps
	}

	// Sort by timestamp
	sorted := make([]DataPoint, len(points))
	copy(sorted, points)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	// Expected interval (assume daily)
	expectedInterval := 24 * time.Hour

	for i := 1; i < len(sorted); i++ {
		gap := sorted[i].Timestamp.Sub(sorted[i-1].Timestamp)
		if gap > expectedInterval*2 {
			gaps = append(gaps, gap)
		}
	}

	return gaps
}

func (d *Detector) findDuplicates(points []DataPoint) []time.Time {
	seen := make(map[int64]bool)
	var duplicates []time.Time

	for _, p := range points {
		key := p.Timestamp.Unix()
		if seen[key] {
			duplicates = append(duplicates, p.Timestamp)
		}
		seen[key] = true
	}

	return duplicates
}

func extractValues(points []DataPoint) []float64 {
	values := make([]float64, len(points))
	for i, p := range points {
		values[i] = p.Value
	}
	return values
}

func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateStdDev(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0
	}
	sumSquares := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	return math.Sqrt(sumSquares / float64(len(values)-1))
}

func calculateLinearTrend(points []DataPoint) (slope, rSquared float64) {
	n := float64(len(points))
	if n < 2 {
		return 0, 0
	}

	// Sort by time
	sorted := make([]DataPoint, len(points))
	copy(sorted, points)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	// Use day index as x
	baseTime := sorted[0].Timestamp
	var sumX, sumY, sumXY, sumX2, sumY2 float64

	for _, p := range sorted {
		x := p.Timestamp.Sub(baseTime).Hours() / 24 // Days
		y := p.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
		sumY2 += y * y
	}

	// Calculate slope
	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return 0, 0
	}
	slope = (n*sumXY - sumX*sumY) / denom

	// Calculate R-squared
	meanY := sumY / n
	ssTotal := sumY2 - n*meanY*meanY
	if ssTotal == 0 {
		return slope, 1
	}

	intercept := (sumY - slope*sumX) / n
	ssResidual := 0.0
	for _, p := range sorted {
		x := p.Timestamp.Sub(baseTime).Hours() / 24
		predicted := slope*x + intercept
		diff := p.Value - predicted
		ssResidual += diff * diff
	}

	rSquared = 1 - (ssResidual / ssTotal)
	return slope, math.Max(0, rSquared)
}

func severityOrder(s Severity) int {
	switch s {
	case SeverityCritical:
		return 0
	case SeverityHigh:
		return 1
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 3
	default:
		return 4
	}
}
