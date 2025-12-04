package aws

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

// roundTripFunc allows injecting a fake HTTP transport.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestAdapterIngest_UsesMockClientAndBuildsActivities(t *testing.T) {
	// Fake AWS response returned by the mock transport.
	mockResp := `{
		"emissionsByService": [{
			"serviceCode": "AmazonEC2",
			"serviceName": "EC2",
			"co2e": 12.34,
			"region": "us-east-1",
			"scope": "Scope2"
		}],
		"totalCO2e": 12.34,
		"unit": "metric_tons",
		"period": {
			"start": "2024-01-01T00:00:00Z",
			"end": "2024-02-01T00:00:00Z"
		}
	}`

	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			// Ensure the signer produced required headers and the request is POSTed.
			if req.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", req.Method)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(mockResp)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	adapter, err := NewAdapter(Config{
		AccessKeyID:     "AKIA_TEST",
		SecretAccessKey: "secret_secret_secret_secret_123",
		Region:          "us-east-1",
		OrgID:           "org-123",
		StartDate:       start,
		EndDate:         end,
		HTTPClient:      client,
	})
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	activities, err := adapter.Ingest(context.Background())
	if err != nil {
		t.Fatalf("ingest returned error: %v", err)
	}

	if len(activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(activities))
	}

	act := activities[0]
	if act.Source != "aws_carbon_footprint" {
		t.Errorf("unexpected source %s", act.Source)
	}
	if act.Category != "cloud_compute_scope2" {
		t.Errorf("expected scope2 category, got %s", act.Category)
	}
	if act.Quantity != 12.34 {
		t.Errorf("expected quantity 12.34, got %f", act.Quantity)
	}
	if act.Metadata["aws_service_code"] != "AmazonEC2" {
		t.Errorf("metadata missing service code, got %s", act.Metadata["aws_service_code"])
	}
	if act.Location != "US-EAST" {
		t.Errorf("expected mapped region US-EAST, got %s", act.Location)
	}
	if act.PeriodStart != start {
		t.Errorf("expected period start %v, got %v", start, act.PeriodStart)
	}
	if act.PeriodEnd != end {
		t.Errorf("expected period end %v, got %v", end, act.PeriodEnd)
	}
}

func TestConvertCURToActivities_NormalizesUnits(t *testing.T) {
	adapter := &Adapter{
		config: Config{
			OrgID: "org-xyz",
		},
	}

	records := []CURRecord{
		{
			LineItemID:          "line-1",
			ServiceCode:         "AmazonS3",
			ServiceName:         "S3",
			Region:              "us-west-2",
			UsageQuantity:       10.5,
			UsageUnit:           "GB-Mo",
			UsageType:           "TimedStorage-ByteHrs",
			UsageStartDate:      "2024-01-01T00:00:00Z",
			UsageEndDate:        "2024-02-01T00:00:00Z",
			BlendedCost:         1.23,
			UnblendedCost:       1.23,
			ProductInstanceType: "standard",
		},
	}

	activities := adapter.ConvertCURToActivities(records)
	if len(activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(activities))
	}

	if activities[0].Unit != "GB-month" {
		t.Errorf("expected normalized unit GB-month, got %s", activities[0].Unit)
	}
	if activities[0].Category != "cloud_storage" {
		t.Errorf("expected category cloud_storage, got %s", activities[0].Category)
	}
	if activities[0].Metadata["aws_usage_type"] == "" {
		t.Errorf("expected usage type metadata to be set")
	}
	if activities[0].OrgID != "org-xyz" {
		t.Errorf("expected org ID to be propagated")
	}
}
