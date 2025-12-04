// Package aws provides an ingestion adapter for AWS Carbon Footprint data.
// It fetches emissions data from AWS Cost and Usage Reports (CUR) and the
// AWS Customer Carbon Footprint Tool API.
package aws

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/example/offgridflow/internal/ingestion"
	"github.com/google/uuid"
)

// =============================================================================
// Configuration
// =============================================================================

// Config holds AWS adapter configuration.
type Config struct {
	// AccessKeyID is the AWS access key ID.
	AccessKeyID string `json:"access_key_id"`

	// SecretAccessKey is the AWS secret access key.
	SecretAccessKey string `json:"-"` // Excluded from JSON

	// Region is the primary AWS region.
	Region string `json:"region"`

	// RoleARN is an optional IAM role ARN for cross-account access.
	RoleARN string `json:"role_arn,omitempty"`

	// AccountID is the AWS account ID for filtering data.
	AccountID string `json:"account_id,omitempty"`

	// OrgID is the OffGridFlow organization ID to associate activities with.
	OrgID string `json:"org_id"`

	// StartDate is the beginning of the date range to fetch (inclusive).
	StartDate time.Time `json:"start_date"`

	// EndDate is the end of the date range to fetch (exclusive).
	EndDate time.Time `json:"end_date"`

	// Optional CUR bucket/prefix for granular ingestion.
	S3Bucket string `json:"s3_bucket,omitempty"`
	S3Prefix string `json:"s3_prefix,omitempty"`

	// HTTPClient allows injecting a custom HTTP client for testing.
	HTTPClient *http.Client `json:"-"`
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	if c.AccessKeyID == "" {
		return fmt.Errorf("aws: access_key_id is required")
	}
	if c.SecretAccessKey == "" {
		return fmt.Errorf("aws: secret_access_key is required")
	}
	if c.Region == "" {
		return fmt.Errorf("aws: region is required")
	}
	if c.OrgID == "" {
		return fmt.Errorf("aws: org_id is required")
	}
	return nil
}

// =============================================================================
// AWS API Response Types
// =============================================================================

// CarbonFootprintSummary represents the AWS Carbon Footprint API response.
type CarbonFootprintSummary struct {
	EmissionsByService []ServiceEmission `json:"emissionsByService"`
	TotalCO2e          float64           `json:"totalCO2e"`
	Unit               string            `json:"unit"` // metric tons
	Period             Period            `json:"period"`
}

// ServiceEmission represents emissions for a single AWS service.
type ServiceEmission struct {
	ServiceCode string  `json:"serviceCode"`
	ServiceName string  `json:"serviceName"`
	CO2e        float64 `json:"co2e"`
	Region      string  `json:"region"`
	Scope       string  `json:"scope"` // Scope1, Scope2, Scope3
}

// Period represents a time period.
type Period struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// CURRecord represents a row from AWS Cost and Usage Report.
type CURRecord struct {
	LineItemID          string  `json:"lineItemId"`
	ServiceCode         string  `json:"serviceCode"`
	ServiceName         string  `json:"serviceName"`
	Region              string  `json:"region"`
	UsageQuantity       float64 `json:"usageQuantity"`
	UsageUnit           string  `json:"usageUnit"`
	UsageType           string  `json:"usageType"`
	UsageStartDate      string  `json:"usageStartDate"`
	UsageEndDate        string  `json:"usageEndDate"`
	BlendedCost         float64 `json:"blendedCost"`
	UnblendedCost       float64 `json:"unblendedCost"`
	ProductInstanceType string  `json:"productInstanceType,omitempty"`
}

// =============================================================================
// Adapter Implementation
// =============================================================================

// Adapter ingests carbon emissions data from AWS.
type Adapter struct {
	config Config
	client *http.Client
}

// NewAdapter creates a new AWS ingestion adapter.
func NewAdapter(cfg Config) (*Adapter, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &Adapter{
		config: cfg,
		client: client,
	}, nil
}

// Ingest fetches carbon emissions data from AWS and returns activities.
func (a *Adapter) Ingest(ctx context.Context) ([]ingestion.Activity, error) {
	var (
		footprint *CarbonFootprintSummary
		err       error
	)

	err = ingestion.WithRetry(ctx, 3, 2*time.Second, func() error {
		footprint, err = a.fetchCarbonFootprint(ctx)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("aws: failed to fetch carbon footprint: %w", err)
	}

	// Convert to activities
	activities := a.convertToActivities(footprint)

	return activities, nil
}

// fetchCarbonFootprint calls the AWS Carbon Footprint API.
// Note: In production, this would use the AWS SDK with proper authentication.
func (a *Adapter) fetchCarbonFootprint(ctx context.Context) (*CarbonFootprintSummary, error) {
	// Build the API endpoint
	endpoint := fmt.Sprintf(
		"https://ce.%s.amazonaws.com/GetCarbonFootprintSummary",
		a.config.Region,
	)

	// Build request body
	requestBody := map[string]interface{}{
		"TimePeriod": map[string]string{
			"Start": a.config.StartDate.Format("2006-01-02"),
			"End":   a.config.EndDate.Format("2006-01-02"),
		},
		"Granularity": "MONTHLY",
	}

	bodyBytes, _ := json.Marshal(requestBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, err
	}

	// Add AWS authentication headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Amz-Target", "AWSInsightsIndexService.GetCarbonFootprintSummary")

	// Add AWS Signature V4 authentication
	if err := a.signRequest(ctx, req); err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("aws: API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result CarbonFootprintSummary
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("aws: failed to decode response: %w", err)
	}

	return &result, nil
}

// signRequest adds AWS Signature V4 headers to the request using static credentials.
func (a *Adapter) signRequest(ctx context.Context, req *http.Request) error {
	provider := awsv2.NewCredentialsCache(credentials.NewStaticCredentialsProvider(a.config.AccessKeyID, a.config.SecretAccessKey, ""))
	creds, err := provider.Retrieve(ctx)
	if err != nil {
		return fmt.Errorf("aws: retrieve credentials: %w", err)
	}
	signer := v4.NewSigner()
	return signer.SignHTTP(ctx, creds, req, "UNSIGNED-PAYLOAD", "ce", a.config.Region, time.Now())
}

// convertToActivities transforms AWS carbon data into OffGridFlow activities.
func (a *Adapter) convertToActivities(footprint *CarbonFootprintSummary) []ingestion.Activity {
	activities := make([]ingestion.Activity, 0, len(footprint.EmissionsByService))
	now := time.Now().UTC()

	for _, emission := range footprint.EmissionsByService {
		// Determine scope from AWS data
		scope := determineScope(emission.Scope)

		// Map AWS region to OffGridFlow location code
		location := mapAWSRegion(emission.Region)

		activity := ingestion.Activity{
			ID:          uuid.NewString(),
			Source:      "aws_carbon_footprint",
			Category:    fmt.Sprintf("cloud_compute_%s", scope),
			Location:    location,
			PeriodStart: footprint.Period.Start,
			PeriodEnd:   footprint.Period.End,
			Quantity:    emission.CO2e,
			Unit:        "tonne", // AWS reports in metric tons CO2e
			OrgID:       a.config.OrgID,
			Metadata: map[string]string{
				"aws_service_code": emission.ServiceCode,
				"aws_service_name": emission.ServiceName,
				"aws_region":       emission.Region,
				"emission_scope":   emission.Scope,
				"data_source":      "aws_carbon_footprint_tool",
			},
			CreatedAt:   now,
			DataQuality: "measured",
			ExternalID:  fmt.Sprintf("aws_%s_%s_%s", a.config.AccountID, emission.ServiceCode, emission.Region),
		}

		activities = append(activities, activity)
	}

	return activities
}

// determineScope normalizes AWS scope designations.
func determineScope(awsScope string) string {
	switch strings.ToLower(awsScope) {
	case "scope1":
		return "scope1"
	case "scope2":
		return "scope2"
	case "scope3":
		return "scope3"
	default:
		return "scope3" // Cloud compute is typically Scope 3 for customers
	}
}

// mapAWSRegion converts AWS region codes to OffGridFlow location codes.
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
		"ca-central-1":   "US-EAST", // Canada uses US-East grid
		"me-south-1":     "MENA",
		"af-south-1":     "AFRICA",
	}

	if location, ok := regionMap[awsRegion]; ok {
		return location
	}
	return "GLOBAL" // Default fallback
}

// =============================================================================
// Cost and Usage Report (CUR) Integration
// =============================================================================

// FetchCURRecords retrieves data from AWS Cost and Usage Reports.
// This requires CUR to be configured in the AWS account.
func (a *Adapter) FetchCURRecords(ctx context.Context, s3Bucket, s3Prefix string) ([]CURRecord, error) {
	if s3Bucket == "" {
		if a.config.S3Bucket == "" {
			return nil, fmt.Errorf("aws: s3 bucket not configured for CUR ingestion")
		}
		s3Bucket = a.config.S3Bucket
	}
	if s3Prefix == "" {
		s3Prefix = a.config.S3Prefix
	}

	cfg, err := awscfg.LoadDefaultConfig(ctx,
		awscfg.WithRegion(a.config.Region),
		awscfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(a.config.AccessKeyID, a.config.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("aws: failed to load config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	listOut, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: &s3Bucket,
		Prefix: awsv2.String(s3Prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("aws: list CUR objects: %w", err)
	}
	if len(listOut.Contents) == 0 {
		return nil, fmt.Errorf("aws: no CUR files found in %s/%s", s3Bucket, s3Prefix)
	}

	var target *types.Object
	for _, obj := range listOut.Contents {
		if strings.HasSuffix(*obj.Key, ".csv") {
			target = &obj
			break
		}
	}
	if target == nil {
		target = &listOut.Contents[0]
	}

	objOut, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s3Bucket,
		Key:    target.Key,
	})
	if err != nil {
		return nil, fmt.Errorf("aws: fetch CUR object: %w", err)
	}
	defer objOut.Body.Close()

	records, err := parseCURCSV(objOut.Body)
	if err != nil {
		return nil, fmt.Errorf("aws: parse CUR csv: %w", err)
	}

	return records, nil
}

// ConvertCURToActivities transforms CUR records into activities.
// This enables more granular emissions tracking by service and resource.
func (a *Adapter) ConvertCURToActivities(records []CURRecord) []ingestion.Activity {
	activities := make([]ingestion.Activity, 0, len(records))
	now := time.Now().UTC()

	for _, record := range records {
		// Parse dates from CUR format
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
			OrgID:       a.config.OrgID,
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

// parseCURCSV parses a CUR CSV stream into CURRecord entries.
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

// categorizeAWSService maps AWS service codes to emission categories.
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

// normalizeAWSUnit converts AWS usage units to standard units.
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
