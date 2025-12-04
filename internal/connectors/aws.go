package connectors

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3API defines the interface for S3 operations
type S3API interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

// CostExplorerAPI defines the interface for Cost Explorer operations
type CostExplorerAPI interface {
	GetCostAndUsage(ctx context.Context, params *costexplorer.GetCostAndUsageInput, optFns ...func(*costexplorer.Options)) (*costexplorer.GetCostAndUsageOutput, error)
}

// AWSConnector handles AWS Cost and Usage Report ingestion
type AWSConnector struct {
	ceClient CostExplorerAPI
	s3Client S3API
	config   ConnectorAWSConfig
}

// ConnectorAWSConfig holds AWS connector configuration
type ConnectorAWSConfig struct {
	Region          string
	RoleARN         string
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	Prefix          string
	EmissionFactors map[string]float64 // region -> kgCO2e/kWh
}

// AWSUsageRecord represents a parsed AWS usage record
type AWSUsageRecord struct {
	Date           time.Time
	AccountID      string
	Service        string
	Region         string
	UsageType      string
	UsageAmount    float64
	Cost           float64
	EmissionsKgCO2 float64
}

// NewAWSConnector creates a new AWS connector with real clients
func NewAWSConnector(ctx context.Context, cfg ConnectorAWSConfig) (*AWSConnector, error) {
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}

	var awsCfg aws.Config
	var err error

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			)),
			config.WithRetryMaxAttempts(3),
			config.WithRetryer(func() aws.Retryer {
				return retry.AddWithMaxAttempts(retry.NewStandard(), 3)
			}),
		)
	} else {
		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
			config.WithRetryMaxAttempts(3),
			config.WithRetryer(func() aws.Retryer {
				return retry.AddWithMaxAttempts(retry.NewStandard(), 3)
			}),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Set default emission factors if not provided
	if cfg.EmissionFactors == nil {
		cfg.EmissionFactors = defaultAWSEmissionFactors()
	}

	return &AWSConnector{
		ceClient: costexplorer.NewFromConfig(awsCfg),
		s3Client: s3.NewFromConfig(awsCfg),
		config:   cfg,
	}, nil
}

// NewAWSConnectorWithClients creates a connector with injected clients (for testing)
func NewAWSConnectorWithClients(cfg ConnectorAWSConfig, ceClient CostExplorerAPI, s3Client S3API) *AWSConnector {
	if cfg.EmissionFactors == nil {
		cfg.EmissionFactors = defaultAWSEmissionFactors()
	}
	return &AWSConnector{
		ceClient: ceClient,
		s3Client: s3Client,
		config:   cfg,
	}
}

// FetchCostAndUsage retrieves cost and usage data using Cost Explorer API with pagination
func (c *AWSConnector) FetchCostAndUsage(ctx context.Context, startDate, endDate time.Time) ([]AWSUsageRecord, error) {
	start := startDate.Format("2006-01-02")
	end := endDate.Format("2006-01-02")

	records := make([]AWSUsageRecord, 0)
	var nextPageToken *string

	for {
		input := &costexplorer.GetCostAndUsageInput{
			TimePeriod: &types.DateInterval{
				Start: aws.String(start),
				End:   aws.String(end),
			},
			Granularity: types.GranularityDaily,
			Metrics:     []string{"UnblendedCost", "UsageQuantity"},
			GroupBy: []types.GroupDefinition{
				{
					Type: types.GroupDefinitionTypeDimension,
					Key:  aws.String("SERVICE"),
				},
				{
					Type: types.GroupDefinitionTypeDimension,
					Key:  aws.String("REGION"),
				},
			},
			NextPageToken: nextPageToken,
		}

		result, err := c.ceClient.GetCostAndUsage(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch cost and usage: %w", err)
		}

		for _, resultByTime := range result.ResultsByTime {
			date, _ := time.Parse("2006-01-02", *resultByTime.TimePeriod.Start)

			for _, group := range resultByTime.Groups {
				service := ""
				region := ""
				if len(group.Keys) >= 2 {
					service = group.Keys[0]
					region = group.Keys[1]
				}

				cost := 0.0
				usage := 0.0

				if metric, ok := group.Metrics["UnblendedCost"]; ok && metric.Amount != nil {
					fmt.Sscanf(*metric.Amount, "%f", &cost)
				}
				if metric, ok := group.Metrics["UsageQuantity"]; ok && metric.Amount != nil {
					fmt.Sscanf(*metric.Amount, "%f", &usage)
				}

				// Calculate emissions based on usage and region
				emissions := c.calculateEmissions(service, region, usage)

				records = append(records, AWSUsageRecord{
					Date:           date,
					AccountID:      c.config.AccountID,
					Service:        service,
					Region:         region,
					UsageAmount:    usage,
					Cost:           cost,
					EmissionsKgCO2: emissions,
				})
			}
		}

		// Check for more pages
		if result.NextPageToken == nil || *result.NextPageToken == "" {
			break
		}
		nextPageToken = result.NextPageToken
	}

	return records, nil
}

// FetchCURFromS3 retrieves and parses Cost and Usage Report from S3 with pagination
func (c *AWSConnector) FetchCURFromS3(ctx context.Context) ([]AWSUsageRecord, error) {
	if c.config.Bucket == "" {
		return nil, fmt.Errorf("S3 bucket not configured")
	}

	var allRecords []AWSUsageRecord
	var continuationToken *string

	// List all CUR CSV files with pagination
	for {
		listInput := &s3.ListObjectsV2Input{
			Bucket:            aws.String(c.config.Bucket),
			Prefix:            aws.String(c.config.Prefix),
			ContinuationToken: continuationToken,
		}

		listOut, err := c.s3Client.ListObjectsV2(ctx, listInput)
		if err != nil {
			return nil, fmt.Errorf("failed to list S3 objects: %w", err)
		}

		// Process CSV files
		for _, obj := range listOut.Contents {
			if !strings.HasSuffix(*obj.Key, ".csv") && !strings.HasSuffix(*obj.Key, ".csv.gz") {
				continue
			}

			records, err := c.fetchAndParseCURFile(ctx, *obj.Key)
			if err != nil {
				// Log error but continue with other files
				continue
			}
			allRecords = append(allRecords, records...)
		}

		if listOut.IsTruncated != nil && !*listOut.IsTruncated {
			break
		}
		continuationToken = listOut.NextContinuationToken
	}

	return allRecords, nil
}

// fetchAndParseCURFile fetches and parses a single CUR file from S3
func (c *AWSConnector) fetchAndParseCURFile(ctx context.Context, key string) ([]AWSUsageRecord, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(c.config.Bucket),
		Key:    aws.String(key),
	}

	result, err := c.s3Client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get CUR from S3: %w", err)
	}
	defer result.Body.Close()

	return c.parseCUR(result.Body)
}

// parseCUR parses the CSV Cost and Usage Report
func (c *AWSConnector) parseCUR(r io.Reader) ([]AWSUsageRecord, error) {
	reader := csv.NewReader(r)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Map column names to indices
	colMap := make(map[string]int)
	for i, col := range header {
		colMap[col] = i
	}

	records := make([]AWSUsageRecord, 0)
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}

		record, err := c.parseRow(row, colMap)
		if err != nil {
			continue // Skip invalid rows
		}

		records = append(records, record)
	}

	return records, nil
}

func (c *AWSConnector) parseRow(row []string, colMap map[string]int) (AWSUsageRecord, error) {
	record := AWSUsageRecord{
		AccountID: c.config.AccountID,
	}

	// Parse date
	if idx, ok := colMap["lineItem/UsageStartDate"]; ok && idx < len(row) {
		date, err := time.Parse("2006-01-02T15:04:05Z", row[idx])
		if err != nil {
			return record, err
		}
		record.Date = date
	}

	// Parse service
	if idx, ok := colMap["lineItem/ProductCode"]; ok && idx < len(row) {
		record.Service = row[idx]
	}

	// Parse region
	if idx, ok := colMap["product/region"]; ok && idx < len(row) {
		record.Region = row[idx]
	}

	// Parse usage type
	if idx, ok := colMap["lineItem/UsageType"]; ok && idx < len(row) {
		record.UsageType = row[idx]
	}

	// Parse usage amount
	if idx, ok := colMap["lineItem/UsageAmount"]; ok && idx < len(row) {
		fmt.Sscanf(row[idx], "%f", &record.UsageAmount)
	}

	// Parse cost
	if idx, ok := colMap["lineItem/UnblendedCost"]; ok && idx < len(row) {
		fmt.Sscanf(row[idx], "%f", &record.Cost)
	}

	// Calculate emissions
	record.EmissionsKgCO2 = c.calculateEmissions(record.Service, record.Region, record.UsageAmount)

	return record, nil
}

func (c *AWSConnector) calculateEmissions(service, region string, usage float64) float64 {
	// For compute services, estimate power consumption
	if strings.Contains(strings.ToLower(service), "ec2") ||
		strings.Contains(strings.ToLower(service), "ecs") ||
		strings.Contains(strings.ToLower(service), "eks") {

		// Estimate: 1 vCPU-hour ≈ 0.025 kWh (rough average)
		estimatedKWh := usage * 0.025

		// Get region emission factor
		factor, ok := c.config.EmissionFactors[region]
		if !ok {
			factor = 0.4 // Default US average
		}

		return estimatedKWh * factor
	}

	// For storage services
	if strings.Contains(strings.ToLower(service), "s3") ||
		strings.Contains(strings.ToLower(service), "ebs") {
		// Estimate: 1 GB-month ≈ 0.00001 kWh
		estimatedKWh := usage * 0.00001

		factor, ok := c.config.EmissionFactors[region]
		if !ok {
			factor = 0.4
		}

		return estimatedKWh * factor
	}

	return 0.0
}

func defaultAWSEmissionFactors() map[string]float64 {
	return map[string]float64{
		"us-east-1":      0.3854, // Virginia (SRVC)
		"us-east-2":      0.5821, // Ohio (RFC)
		"us-west-1":      0.2451, // California (CAMX)
		"us-west-2":      0.2883, // Oregon (NWPP)
		"eu-west-1":      0.2929, // Ireland
		"eu-central-1":   0.3380, // Frankfurt
		"ap-southeast-1": 0.4990, // Singapore
		"ap-northeast-1": 0.4630, // Tokyo
		"ap-south-1":     0.7080, // Mumbai
		"sa-east-1":      0.0820, // São Paulo
	}
}

// Close cleans up resources
func (c *AWSConnector) Close() error {
	return nil
}
