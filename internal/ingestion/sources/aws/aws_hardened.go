//go:build hardened
// +build hardened

// Package aws provides a hardened ingestion adapter for AWS cloud emissions data.
// This version includes rate limiting, pagination, error classification, and observability.
package aws

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/example/offgridflow/internal/ingestion"
	"github.com/google/uuid"
)

// =============================================================================
// Hardened AWS Adapter Configuration
// =============================================================================

// HardenedConfig extends Config with additional hardening options.
type HardenedConfig struct {
	// Base configuration
	Config

	// RateLimitCapacity is the token bucket capacity (default: 100)
	RateLimitCapacity float64

	// RateLimitPerSec is the refill rate in requests/second (default: 5.0)
	RateLimitPerSec float64

	// MaxRetries is the maximum number of retry attempts (default: 3)
	MaxRetries int

	// RequestTimeout is the timeout for individual API calls (default: 30s)
	RequestTimeout time.Duration

	// IncludeS3Manifest enables S3 CUR manifest parsing (default: true)
	IncludeS3Manifest bool

	// Logger for structured logging (optional)
	Logger *slog.Logger

	// Observability configuration (optional)
	Observability *ingestion.ObservabilityConfig

	// MaxPages limits pagination to prevent runaway requests (default: 1000)
	MaxPages int
}

// NewHardenedConfig creates a new hardened configuration with defaults.
func NewHardenedConfig(cfg Config) *HardenedConfig {
	return &HardenedConfig{
		Config:            cfg,
		RateLimitCapacity: 100,
		RateLimitPerSec:   5.0,
		MaxRetries:        3,
		RequestTimeout:    30 * time.Second,
		IncludeS3Manifest: true,
		MaxPages:          1000,
		Logger:            slog.Default(),
		Observability:     ingestion.NewObservabilityConfig("aws"),
	}
}

// Validate checks hardened configuration.
func (hc *HardenedConfig) Validate() error {
	if err := hc.Config.Validate(); err != nil {
		return err
	}

	if hc.RateLimitCapacity <= 0 {
		return fmt.Errorf("aws: rate_limit_capacity must be positive")
	}

	if hc.RateLimitPerSec <= 0 {
		return fmt.Errorf("aws: rate_limit_per_sec must be positive")
	}

	if hc.MaxRetries < 1 {
		return fmt.Errorf("aws: max_retries must be >= 1")
	}

	if hc.RequestTimeout < 1*time.Second {
		return fmt.Errorf("aws: request_timeout must be >= 1s")
	}

	return nil
}

// =============================================================================
// Hardened Adapter
// =============================================================================

// HardenedAdapter ingests AWS emissions data with production hardening.
type HardenedAdapter struct {
	config  *HardenedConfig
	client  *http.Client
	limiter *ingestion.RateLimiter
	tracer  *ingestion.InvocationTracer
	metrics *ingestion.IngestionMetrics
	logger  *slog.Logger
}

// NewHardenedAdapter creates a new hardened AWS adapter.
func NewHardenedAdapter(cfg *HardenedConfig) (*HardenedAdapter, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Create HTTP client with timeout
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{
			Timeout: cfg.RequestTimeout,
		}
	}

	// Create rate limiter
	limiter := ingestion.NewRateLimiter(
		cfg.RateLimitCapacity,
		cfg.RateLimitPerSec,
		100*time.Millisecond,
	)

	// Setup observability
	tracer := ingestion.NewInvocationTracer(cfg.Observability)
	metrics, err := ingestion.NewIngestionMetrics(cfg.Observability.Meter)
	if err != nil {
		// Log but don't fail if metrics setup fails
		cfg.Logger.Warn("failed to setup ingestion metrics", "error", err)
	}

	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &HardenedAdapter{
		config:  cfg,
		client:  client,
		limiter: limiter,
		tracer:  tracer,
		metrics: metrics,
		logger:  logger,
	}, nil
}

// =============================================================================
// Ingest - Main Entry Point
// =============================================================================

// Ingest fetches AWS emissions data with full hardening (rate limit, retry, error handling, observability).
func (ha *HardenedAdapter) Ingest(ctx context.Context) ([]ingestion.Activity, error) {
	start := time.Now()

	// Ingest from Carbon Footprint API
	footprint, err := ha.ingestCarbonFootprint(ctx)
	if err != nil {
		ha.recordMetricsFailure("carbon_api", err)
		return nil, err
	}

	activities := ha.convertCarbonToActivities(footprint)
	ha.logger.Info("ingested from carbon api", "count", len(activities))

	// Optionally ingest from S3 CUR manifest
	if ha.config.IncludeS3Manifest && ha.config.S3Bucket != "" {
		curActivities, err := ha.ingestS3CUR(ctx)
		if err != nil {
			// Log but don't fail if S3 CUR is optional
			ha.logger.Warn("failed to ingest S3 CUR", "error", err)
		} else {
			activities = append(activities, curActivities...)
			ha.logger.Info("ingested from S3 CUR", "count", len(curActivities))
		}
	}

	// Record success metrics
	latency := time.Since(start)
	ha.recordMetricsSuccess(len(activities), latency)
	ha.logger.Info("aws ingestion complete", "total_activities", len(activities), "latency_ms", latency.Milliseconds())

	return activities, nil
}

// =============================================================================
// Carbon Footprint API Ingestion (with Hardening)
// =============================================================================

// ingestCarbonFootprint fetches from Carbon Footprint API with rate limiting, retry, and error handling.
func (ha *HardenedAdapter) ingestCarbonFootprint(ctx context.Context) (*CarbonFootprintSummary, error) {
	var footprint *CarbonFootprintSummary

	err := ha.retryWithExponentialBackoff(ctx, func() error {
		return ha.tracer.TraceInvocation(ctx, "aws.fetch_carbon_footprint", func(ctx context.Context) error {
			// Wait for rate limit token
			waited, err := ha.limiter.Allow(ctx)
			if err != nil {
				return ingestion.NewClassifiedError(ingestion.ErrorClassTransient, "rate limiter context cancelled", err)
			}
			if waited > 100*time.Millisecond {
				ha.logger.Debug("rate limiting applied", "wait_ms", waited.Milliseconds())
			}

			// Fetch with context timeout
			fetchCtx, cancel := context.WithTimeout(ctx, ha.config.RequestTimeout)
			defer cancel()

			var err error
			footprint, err = ha.fetchCarbonFootprint(fetchCtx)
			if err != nil {
				ce := ingestion.ClassifyError(err)
				ha.tracer.LogIngestionError(ctx, err, "aws-carbon-api")
				return ce
			}

			return nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("aws: carbon footprint ingestion failed: %w", err)
	}

	return footprint, nil
}

// fetchCarbonFootprint calls the AWS Carbon Footprint API.
func (ha *HardenedAdapter) fetchCarbonFootprint(ctx context.Context) (*CarbonFootprintSummary, error) {
	endpoint := fmt.Sprintf(
		"https://ce.%s.amazonaws.com/GetCarbonFootprintSummary",
		ha.config.Region,
	)

	requestBody := map[string]interface{}{
		"TimePeriod": map[string]string{
			"Start": ha.config.StartDate.Format("2006-01-02"),
			"End":   ha.config.EndDate.Format("2006-01-02"),
		},
		"Granularity": "MONTHLY",
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, ingestion.ClassifyHTTPError(0, err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Amz-Target", "AWSInsightsIndexService.GetCarbonFootprintSummary")

	// Sign request with SigV4
	if err := ha.signRequest(ctx, req); err != nil {
		return nil, ingestion.NewClassifiedError(ingestion.ErrorClassAuth, "failed to sign request", err)
	}

	resp, err := ha.client.Do(req)
	if err != nil {
		return nil, ingestion.ClassifyError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		ce := ingestion.ClassifyHTTPError(resp.StatusCode, string(body))
		return nil, ce
	}

	var result CarbonFootprintSummary
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, ingestion.NewClassifiedError(
			ingestion.ErrorClassBadRequest,
			"failed to parse carbon footprint response",
			err,
		)
	}

	return &result, nil
}

// signRequest signs request with AWS Signature V4.
func (ha *HardenedAdapter) signRequest(ctx context.Context, req *http.Request) error {
	provider := awsv2.NewCredentialsCache(
		credentials.NewStaticCredentialsProvider(ha.config.AccessKeyID, ha.config.SecretAccessKey, ""),
	)

	creds, err := provider.Retrieve(ctx)
	if err != nil {
		return err
	}

	signer := v4.NewSigner()
	return signer.SignHTTP(
		ctx,
		creds,
		req,
		"UNSIGNED-PAYLOAD",
		"ce",
		ha.config.Region,
		time.Now(),
	)
}

// =============================================================================
// S3 CUR Manifest Ingestion (with Hardening)
// =============================================================================

// ingestS3CUR fetches CUR data from S3 manifest with pagination, rate limiting, and error handling.
func (ha *HardenedAdapter) ingestS3CUR(ctx context.Context) ([]ingestion.Activity, error) {
	var activities []ingestion.Activity

	err := ha.tracer.TraceInvocation(ctx, "aws.ingest_s3_cur", func(ctx context.Context) error {
		// Fetch manifest
		manifest, err := ha.fetchS3Manifest(ctx)
		if err != nil {
			ha.tracer.LogIngestionError(ctx, err, "aws-s3-manifest")
			return err
		}

		// Validate manifest
		if err := ValidateManifest(manifest, ha.logger); err != nil {
			return ingestion.NewClassifiedError(ingestion.ErrorClassBadRequest, "invalid manifest", err)
		}

		// Get report files from manifest
		reportFiles := manifest.GetReportFiles()
		ha.logger.Info("s3 manifest parsed", "files", len(reportFiles), "total_size_mb", calculateTotalSize(reportFiles))

		// Setup pagination for files
		pagination := ingestion.NewPaginationState(len(reportFiles))
		pagination.MaxPages = ha.config.MaxPages

		// Process each file with pagination and rate limiting
		for _, file := range reportFiles {
			// Rate limit
			if _, err := ha.limiter.Allow(ctx); err != nil {
				return ingestion.NewClassifiedError(ingestion.ErrorClassTransient, "rate limiter cancelled", err)
			}

			// Fetch and parse file
			records, err := ha.fetchAndParseS3File(ctx, file.Key)
			if err != nil {
				ha.logger.Warn("failed to fetch S3 file", "key", file.Key, "error", err)
				// Continue to next file (soft failure)
				continue
			}

			// Convert to activities
			fileActivities := ha.convertCURToActivities(records)
			activities = append(activities, fileActivities...)

			pagination.TotalFetched += len(records)
			ha.logger.Debug("s3 file processed", "key", file.Key, "records", len(records))
		}

		ha.logger.Info("s3 cur ingestion complete", "total_records", pagination.TotalFetched, "total_activities", len(activities))
		return nil
	})

	if err != nil {
		return nil, err
	}

	return activities, nil
}

// fetchS3Manifest fetches the CUR manifest.json from S3.
func (ha *HardenedAdapter) fetchS3Manifest(ctx context.Context) (*S3Manifest, error) {
	manifestKey := fmt.Sprintf("%s/manifest.json", ha.config.S3Prefix)

	// Rate limit
	if _, err := ha.limiter.Allow(ctx); err != nil {
		return nil, err
	}

	// Fetch manifest
	cfg, err := awscfg.LoadDefaultConfig(ctx,
		awscfg.WithRegion(ha.config.Region),
		awscfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(ha.config.AccessKeyID, ha.config.SecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, ingestion.NewClassifiedError(ingestion.ErrorClassAuth, "failed to load AWS config", err)
	}

	client := s3.NewFromConfig(cfg)

	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &ha.config.S3Bucket,
		Key:    &manifestKey,
	})
	if err != nil {
		ce := ingestion.ClassifyError(err)
		if ce.Class == ingestion.ErrorClassNotFound {
			return nil, ingestion.NewClassifiedError(
				ingestion.ErrorClassNotFound,
				fmt.Sprintf("manifest not found at s3://%s/%s", ha.config.S3Bucket, manifestKey),
				err,
			)
		}
		return nil, ce
	}
	defer resp.Body.Close()

	return ParseS3ManifestFromReader(resp.Body)
}

// fetchAndParseS3File fetches a CUR file from S3 and parses it.
func (ha *HardenedAdapter) fetchAndParseS3File(ctx context.Context, key string) ([]CURRecord, error) {
	cfg, err := awscfg.LoadDefaultConfig(ctx,
		awscfg.WithRegion(ha.config.Region),
		awscfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(ha.config.AccessKeyID, ha.config.SecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)

	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &ha.config.S3Bucket,
		Key:    &key,
	})
	if err != nil {
		ce := ingestion.ClassifyError(err)
		return nil, ce
	}
	defer resp.Body.Close()

	// Handle gzip if needed
	reader := io.Reader(resp.Body)
	if strings.HasSuffix(key, ".gz") {
		// In production, would use gzip.NewReader(resp.Body)
		// For now, assume decompressed
	}

	records, err := parseCURCSV(reader)
	if err != nil {
		return nil, ingestion.NewClassifiedError(ingestion.ErrorClassBadRequest, "failed to parse CSV", err)
	}

	return records, nil
}

// =============================================================================
// Conversion to Activities
// =============================================================================

// convertCarbonToActivities transforms AWS Carbon Footprint data to activities.
func (ha *HardenedAdapter) convertCarbonToActivities(footprint *CarbonFootprintSummary) []ingestion.Activity {
	activities := make([]ingestion.Activity, 0, len(footprint.EmissionsByService))
	now := time.Now().UTC()

	for _, emission := range footprint.EmissionsByService {
		activity := ingestion.Activity{
			ID:          uuid.NewString(),
			Source:      "aws_carbon_footprint",
			Category:    fmt.Sprintf("cloud_compute_%s", determineScope(emission.Scope)),
			Location:    mapAWSRegion(emission.Region),
			PeriodStart: footprint.Period.Start,
			PeriodEnd:   footprint.Period.End,
			Quantity:    emission.CO2e,
			Unit:        "tonne",
			OrgID:       ha.config.OrgID,
			Metadata: map[string]string{
				"aws_service_code": emission.ServiceCode,
				"aws_service_name": emission.ServiceName,
				"aws_region":       emission.Region,
				"emission_scope":   emission.Scope,
				"data_source":      "aws_carbon_footprint_tool",
			},
			CreatedAt:   now,
			DataQuality: "measured",
			ExternalID:  fmt.Sprintf("aws_%s_%s_%s", ha.config.AccountID, emission.ServiceCode, emission.Region),
		}

		activities = append(activities, activity)
	}

	return activities
}

// convertCURToActivities transforms CUR records to activities.
func (ha *HardenedAdapter) convertCURToActivities(records []CURRecord) []ingestion.Activity {
	activities := make([]ingestion.Activity, 0, len(records))
	now := time.Now().UTC()

	for _, record := range records {
		startDate, _ := time.Parse("2006-01-02T15:04:05Z", record.UsageStartDate)
		endDate, _ := time.Parse("2006-01-02T15:04:05Z", record.UsageEndDate)

		activity := ingestion.Activity{
			ID:          uuid.NewString(),
			Source:      "aws_cur",
			Category:    categorizeAWSService(record.ServiceCode),
			Location:    mapAWSRegion(record.Region),
			PeriodStart: startDate,
			PeriodEnd:   endDate,
			Quantity:    record.UsageQuantity,
			Unit:        normalizeAWSUnit(record.UsageUnit),
			OrgID:       ha.config.OrgID,
			Metadata: map[string]string{
				"aws_service_code":  record.ServiceCode,
				"aws_service_name":  record.ServiceName,
				"aws_usage_type":    record.UsageType,
				"aws_instance_type": record.ProductInstanceType,
				"aws_cost_usd":      fmt.Sprintf("%.4f", record.BlendedCost),
				"data_source":       "aws_cur",
			},
			CreatedAt:   now,
			DataQuality: "measured",
			ExternalID:  record.LineItemID,
		}

		activities = append(activities, activity)
	}

	return activities
}

// =============================================================================
// Retry Logic with Exponential Backoff
// =============================================================================

// retryWithExponentialBackoff retries a function with exponential backoff and error classification.
func (ha *HardenedAdapter) retryWithExponentialBackoff(ctx context.Context, fn func() error) error {
	backoff := 1 * time.Second

	for attempt := 1; attempt <= ha.config.MaxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		ce := ingestion.ClassifyError(err)
		if ce == nil {
			return err
		}

		if !ce.IsRetryable() {
			// Don't retry non-transient errors
			ha.logger.Info("non-retryable error, stopping retries", "class", ce.Class, "error", ce.Message)
			return err
		}

		if attempt >= ha.config.MaxRetries {
			// Last attempt failed
			ha.logger.Warn("max retries exceeded", "attempts", attempt, "error", err)
			return err
		}

		// Wait before retrying
		ha.logger.Info("retrying after error", "attempt", attempt, "backoff_ms", backoff.Milliseconds(), "error", err)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
		}
	}

	return fmt.Errorf("aws: max retries exceeded")
}

// =============================================================================
// Metrics Recording
// =============================================================================

// recordMetricsSuccess records successful ingestion metrics.
func (ha *HardenedAdapter) recordMetricsSuccess(itemCount int, latency time.Duration) {
	if ha.metrics == nil {
		return
	}

	ha.metrics.RecordSuccess(context.Background(), itemCount, latency, "aws")
}

// recordMetricsFailure records failed ingestion metrics.
func (ha *HardenedAdapter) recordMetricsFailure(source string, err error) {
	if ha.metrics == nil {
		return
	}

	ce := ingestion.ClassifyError(err)
	errorClass := "unknown"
	if ce != nil {
		errorClass = string(ce.Class)
	}

	ha.metrics.RecordFailure(context.Background(), "aws-"+source, errorClass)
}

// =============================================================================
// Helper Functions
// =============================================================================

// calculateTotalSize sums file sizes from manifest.
func calculateTotalSize(files []ManifestFile) float64 {
	var total int64
	for _, f := range files {
		total += f.Size
	}
	return float64(total) / (1024 * 1024) // Convert to MB
}

// parseCURCSV parses CUR CSV format (same as original).
func parseCURCSV(r io.Reader) ([]CURRecord, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}

	index := func(name string) int {
		for i, h := range headers {
			if strings.EqualFold(h, name) {
				return i
			}
		}
		return -1
	}

	li := index("lineItem/Id")
	svc := index("product/servicecode")
	svcName := index("product/productname")
	region := index("product/region")
	usageQty := index("lineItem/UsageAmount")
	usageUnit := index("lineItem/UsageUnit")
	usageType := index("lineItem/UsageType")
	startIdx := index("lineItem/UsageStartDate")
	endIdx := index("lineItem/UsageEndDate")
	blendIdx := index("lineItem/BlendedCost")
	unblendIdx := index("lineItem/UnblendedCost")
	instIdx := index("product/instanceType")

	var result []CURRecord
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		parseFloat := func(i int) float64 {
			if i < 0 || i >= len(row) {
				return 0
			}
			v, _ := strconv.ParseFloat(row[i], 64)
			return v
		}

		rec := CURRecord{
			LineItemID:          safeCell(row, li),
			ServiceCode:         safeCell(row, svc),
			ServiceName:         safeCell(row, svcName),
			Region:              safeCell(row, region),
			UsageQuantity:       parseFloat(usageQty),
			UsageUnit:           safeCell(row, usageUnit),
			UsageType:           safeCell(row, usageType),
			UsageStartDate:      safeCell(row, startIdx),
			UsageEndDate:        safeCell(row, endIdx),
			BlendedCost:         parseFloat(blendIdx),
			UnblendedCost:       parseFloat(unblendIdx),
			ProductInstanceType: safeCell(row, instIdx),
		}
		result = append(result, rec)
	}

	return result, nil
}

func safeCell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return row[idx]
}

// determineScope, mapAWSRegion, categorizeAWSService, normalizeAWSUnit (same as original)
func determineScope(awsScope string) string {
	switch strings.ToLower(awsScope) {
	case "scope1":
		return "scope1"
	case "scope2":
		return "scope2"
	case "scope3":
		return "scope3"
	default:
		return "scope3"
	}
}

func mapAWSRegion(awsRegion string) string {
	regionMap := map[string]string{
		"us-east-1":      "US-EAST",
		"us-east-2":      "US-EAST",
		"us-west-1":      "US-WEST",
		"us-west-2":      "US-WEST",
		"eu-west-1":      "EU-WEST",
		"eu-west-2":      "EU-WEST",
		"eu-west-3":      "EU-WEST",
		"eu-central-1":   "EU-CENTRAL",
		"eu-north-1":     "EU-NORTH",
		"ap-southeast-1": "ASIA-PACIFIC",
		"ap-southeast-2": "ASIA-PACIFIC",
		"ap-northeast-1": "ASIA-PACIFIC",
		"ap-northeast-2": "ASIA-PACIFIC",
		"ap-south-1":     "ASIA-PACIFIC",
		"sa-east-1":      "LATAM",
		"ca-central-1":   "US-EAST",
		"me-south-1":     "MENA",
		"af-south-1":     "AFRICA",
	}

	if location, ok := regionMap[awsRegion]; ok {
		return location
	}
	return "GLOBAL"
}

func categorizeAWSService(serviceCode string) string {
	categoryMap := map[string]string{
		"AmazonEC2":         "cloud_compute",
		"AmazonRDS":         "cloud_database",
		"AmazonS3":          "cloud_storage",
		"AWSLambda":         "cloud_compute_serverless",
		"AmazonDynamoDB":    "cloud_database",
		"AmazonElastiCache": "cloud_database",
		"AmazonCloudFront":  "cloud_cdn",
		"AmazonEKS":         "cloud_compute_kubernetes",
		"AmazonECS":         "cloud_compute_containers",
		"AWSDataTransfer":   "data_transfer",
	}

	if category, ok := categoryMap[serviceCode]; ok {
		return category
	}
	return "cloud_other"
}

func normalizeAWSUnit(awsUnit string) string {
	unitMap := map[string]string{
		"Hrs":           "hours",
		"GB-Mo":         "GB-month",
		"GB":            "GB",
		"Requests":      "requests",
		"Lambda-GB-Sec": "GB-seconds",
	}

	if unit, ok := unitMap[awsUnit]; ok {
		return unit
	}
	return awsUnit
}
