package connectors

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	cetypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Mock Cost Explorer Client
type mockCostExplorerClient struct {
	output *costexplorer.GetCostAndUsageOutput
	err    error
}

func (m *mockCostExplorerClient) GetCostAndUsage(ctx context.Context, params *costexplorer.GetCostAndUsageInput, optFns ...func(*costexplorer.Options)) (*costexplorer.GetCostAndUsageOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.output, nil
}

// Mock S3 Client
type mockS3Client struct {
	listOutput *s3.ListObjectsV2Output
	getOutput  *s3.GetObjectOutput
	err        error
}

func (m *mockS3Client) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.listOutput, nil
}

func (m *mockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.getOutput, nil
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func TestAWSConnector_FetchCostAndUsage(t *testing.T) {
	ctx := context.Background()

	t.Run("successful fetch with pagination", func(t *testing.T) {
		mockCE := &mockCostExplorerClient{
			output: &costexplorer.GetCostAndUsageOutput{
				ResultsByTime: []cetypes.ResultByTime{
					{
						TimePeriod: &cetypes.DateInterval{
							Start: aws.String("2024-01-01"),
							End:   aws.String("2024-01-02"),
						},
						Groups: []cetypes.Group{
							{
								Keys: []string{"AmazonEC2", "us-east-1"},
								Metrics: map[string]cetypes.MetricValue{
									"UnblendedCost": {
										Amount: aws.String("100.50"),
										Unit:   aws.String("USD"),
									},
									"UsageQuantity": {
										Amount: aws.String("24.0"),
										Unit:   aws.String("Hrs"),
									},
								},
							},
						},
					},
				},
				NextPageToken: nil,
			},
		}

		cfg := ConnectorAWSConfig{
			Region:    "us-east-1",
			AccountID: "123456789012",
		}

		connector := NewAWSConnectorWithClients(cfg, mockCE, nil)

		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

		records, err := connector.FetchCostAndUsage(ctx, start, end)
		if err != nil {
			t.Fatalf("FetchCostAndUsage failed: %v", err)
		}

		if len(records) != 1 {
			t.Fatalf("expected 1 record, got %d", len(records))
		}

		record := records[0]
		if record.Service != "AmazonEC2" {
			t.Errorf("expected service AmazonEC2, got %s", record.Service)
		}
		if record.Region != "us-east-1" {
			t.Errorf("expected region us-east-1, got %s", record.Region)
		}
		if record.Cost != 100.50 {
			t.Errorf("expected cost 100.50, got %f", record.Cost)
		}
		if record.UsageAmount != 24.0 {
			t.Errorf("expected usage 24.0, got %f", record.UsageAmount)
		}
	})

	t.Run("error handling", func(t *testing.T) {
		mockCE := &mockCostExplorerClient{
			err: context.DeadlineExceeded,
		}

		cfg := ConnectorAWSConfig{
			Region:    "us-east-1",
			AccountID: "123456789012",
		}

		connector := NewAWSConnectorWithClients(cfg, mockCE, nil)

		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

		_, err := connector.FetchCostAndUsage(ctx, start, end)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestAWSConnector_FetchCURFromS3(t *testing.T) {
	ctx := context.Background()

	t.Run("successful CUR fetch", func(t *testing.T) {
		csvContent := `lineItem/UsageStartDate,lineItem/ProductCode,product/region,lineItem/UsageType,lineItem/UsageAmount,lineItem/UnblendedCost
2024-01-01T00:00:00Z,AmazonEC2,us-east-1,BoxUsage:t3.micro,24.0,10.50`

		mockS3 := &mockS3Client{
			listOutput: &s3.ListObjectsV2Output{
				Contents: []types.Object{
					{
						Key: aws.String("cur/report-001.csv"),
					},
				},
				IsTruncated: aws.Bool(false),
			},
			getOutput: &s3.GetObjectOutput{
				Body: nopCloser{strings.NewReader(csvContent)},
			},
		}

		cfg := ConnectorAWSConfig{
			Region:    "us-east-1",
			AccountID: "123456789012",
			Bucket:    "my-cur-bucket",
			Prefix:    "cur/",
		}

		connector := NewAWSConnectorWithClients(cfg, nil, mockS3)

		records, err := connector.FetchCURFromS3(ctx)
		if err != nil {
			t.Fatalf("FetchCURFromS3 failed: %v", err)
		}

		if len(records) != 1 {
			t.Fatalf("expected 1 record, got %d", len(records))
		}

		record := records[0]
		if record.Service != "AmazonEC2" {
			t.Errorf("expected service AmazonEC2, got %s", record.Service)
		}
		if record.Region != "us-east-1" {
			t.Errorf("expected region us-east-1, got %s", record.Region)
		}
	})

	t.Run("error when bucket not configured", func(t *testing.T) {
		cfg := ConnectorAWSConfig{
			Region:    "us-east-1",
			AccountID: "123456789012",
		}

		connector := NewAWSConnectorWithClients(cfg, nil, nil)

		_, err := connector.FetchCURFromS3(ctx)
		if err == nil {
			t.Fatal("expected error for missing bucket, got nil")
		}
	})
}

func TestAWSConnector_CalculateEmissions(t *testing.T) {
	cfg := ConnectorAWSConfig{
		Region:    "us-east-1",
		AccountID: "123456789012",
	}

	connector := NewAWSConnectorWithClients(cfg, nil, nil)

	tests := []struct {
		name     string
		service  string
		region   string
		usage    float64
		expected float64
	}{
		{
			name:     "EC2 compute",
			service:  "AmazonEC2",
			region:   "us-east-1",
			usage:    100.0, // 100 hours
			expected: 100.0 * 0.025 * 0.3854, // hours * kWh/hour * factor
		},
		{
			name:     "S3 storage",
			service:  "AmazonS3",
			region:   "us-west-1",
			usage:    1000000.0, // GB-month
			expected: 1000000.0 * 0.00001 * 0.2451,
		},
		{
			name:     "Unknown service",
			service:  "AmazonUnknown",
			region:   "us-east-1",
			usage:    100.0,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emissions := connector.calculateEmissions(tt.service, tt.region, tt.usage)
			// Allow small floating point differences
			if diff := emissions - tt.expected; diff > 0.001 || diff < -0.001 {
				t.Errorf("expected emissions %.6f, got %.6f", tt.expected, emissions)
			}
		})
	}
}
