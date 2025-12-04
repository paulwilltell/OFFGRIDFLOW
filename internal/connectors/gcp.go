package connectors

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GCPConnector handles Google Cloud Platform carbon footprint data ingestion
type GCPConnector struct {
	client *bigquery.Client
	config GCPConfig
}

// GCPConfig holds GCP connector configuration
type GCPConfig struct {
	ProjectID        string
	BillingAccountID string
	DatasetID        string
	TableID          string
	CredentialsJSON  string // Service account JSON
	EmissionFactors  map[string]float64
}

// GCPUsageRecord represents parsed GCP usage data
type GCPUsageRecord struct {
	Date           time.Time
	ProjectID      string
	ProjectName    string
	Service        string
	SKU            string
	Region         string
	UsageAmount    float64
	UsageUnit      string
	Cost           float64
	EmissionsKgCO2 float64
}

// GCPCarbonRecord represents data from Carbon Footprint export
type GCPCarbonRecord struct {
	Date             time.Time
	ProjectID        string
	ProjectName      string
	Service          string
	Region           string
	CarbonKgCO2e     float64
	LocationBasedCO2 float64
	MarketBasedCO2   float64
}

// NewGCPConnector creates a new GCP connector
func NewGCPConnector(ctx context.Context, cfg GCPConfig) (*GCPConnector, error) {
	if cfg.EmissionFactors == nil {
		cfg.EmissionFactors = defaultGCPEmissionFactors()
	}

	var client *bigquery.Client
	var err error

	if cfg.CredentialsJSON != "" {
		client, err = bigquery.NewClient(ctx, cfg.ProjectID, option.WithCredentialsJSON([]byte(cfg.CredentialsJSON)))
	} else {
		client, err = bigquery.NewClient(ctx, cfg.ProjectID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create BigQuery client: %w", err)
	}

	return &GCPConnector{
		client: client,
		config: cfg,
	}, nil
}

// FetchCarbonFootprint retrieves carbon footprint data from BigQuery Carbon Footprint export
func (c *GCPConnector) FetchCarbonFootprint(ctx context.Context, startDate, endDate time.Time) ([]GCPCarbonRecord, error) {
	query := fmt.Sprintf(`
		SELECT
			usage_month as date,
			project.id as project_id,
			project.name as project_name,
			service.description as service,
			location.region as region,
			carbon_footprint_total_kgCO2e.amount as carbon_kg_co2e,
			carbon_footprint_location_based_kgCO2e.amount as location_based_co2,
			carbon_footprint_market_based_kgCO2e.amount as market_based_co2
		FROM
			%s.%s.%s
		WHERE
			usage_month >= @start_date
			AND usage_month <= @end_date
		ORDER BY
			usage_month DESC
	`, c.config.ProjectID, c.config.DatasetID, c.config.TableID)

	q := c.client.Query(query)
	q.Parameters = []bigquery.QueryParameter{
		{
			Name:  "start_date",
			Value: startDate.Format("2006-01-02"),
		},
		{
			Name:  "end_date",
			Value: endDate.Format("2006-01-02"),
		},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	records := make([]GCPCarbonRecord, 0)
	for {
		var row struct {
			Date         bigquery.NullDate    `bigquery:"date"`
			ProjectID    bigquery.NullString  `bigquery:"project_id"`
			ProjectName  bigquery.NullString  `bigquery:"project_name"`
			Service      bigquery.NullString  `bigquery:"service"`
			Region       bigquery.NullString  `bigquery:"region"`
			CarbonKgCO2e bigquery.NullFloat64 `bigquery:"carbon_kg_co2e"`
			LocationCO2  bigquery.NullFloat64 `bigquery:"location_based_co2"`
			MarketCO2    bigquery.NullFloat64 `bigquery:"market_based_co2"`
		}

		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		record := GCPCarbonRecord{}

		if row.Date.Valid {
			// Convert civil.Date to time.Time
			d := row.Date.Date
			record.Date = time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
		}
		if row.ProjectID.Valid {
			record.ProjectID = row.ProjectID.StringVal
		}
		if row.ProjectName.Valid {
			record.ProjectName = row.ProjectName.StringVal
		}
		if row.Service.Valid {
			record.Service = row.Service.StringVal
		}
		if row.Region.Valid {
			record.Region = row.Region.StringVal
		}
		if row.CarbonKgCO2e.Valid {
			record.CarbonKgCO2e = row.CarbonKgCO2e.Float64
		}
		if row.LocationCO2.Valid {
			record.LocationBasedCO2 = row.LocationCO2.Float64
		}
		if row.MarketCO2.Valid {
			record.MarketBasedCO2 = row.MarketCO2.Float64
		}

		records = append(records, record)
	}

	return records, nil
}

// FetchBillingData retrieves billing data and estimates emissions
func (c *GCPConnector) FetchBillingData(ctx context.Context, startDate, endDate time.Time) ([]GCPUsageRecord, error) {
	query := fmt.Sprintf(`
		SELECT
			DATE(usage_start_time) as date,
			project.id as project_id,
			project.name as project_name,
			service.description as service,
			sku.description as sku,
			location.region as region,
			SUM(usage.amount) as usage_amount,
			usage.unit as usage_unit,
			SUM(cost) as cost
		FROM
			%s.%s.gcp_billing_export_v1_%s
		WHERE
			DATE(usage_start_time) >= @start_date
			AND DATE(usage_start_time) <= @end_date
		GROUP BY
			date, project_id, project_name, service, sku, region, usage_unit
		ORDER BY
			date DESC
	`, c.config.ProjectID, c.config.DatasetID, c.config.BillingAccountID)

	q := c.client.Query(query)
	q.Parameters = []bigquery.QueryParameter{
		{
			Name:  "start_date",
			Value: startDate.Format("2006-01-02"),
		},
		{
			Name:  "end_date",
			Value: endDate.Format("2006-01-02"),
		},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	records := make([]GCPUsageRecord, 0)
	for {
		var row struct {
			Date        bigquery.NullDate    `bigquery:"date"`
			ProjectID   bigquery.NullString  `bigquery:"project_id"`
			ProjectName bigquery.NullString  `bigquery:"project_name"`
			Service     bigquery.NullString  `bigquery:"service"`
			SKU         bigquery.NullString  `bigquery:"sku"`
			Region      bigquery.NullString  `bigquery:"region"`
			UsageAmount bigquery.NullFloat64 `bigquery:"usage_amount"`
			UsageUnit   bigquery.NullString  `bigquery:"usage_unit"`
			Cost        bigquery.NullFloat64 `bigquery:"cost"`
		}

		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		record := GCPUsageRecord{}

		if row.Date.Valid {
			// Convert civil.Date to time.Time
			d := row.Date.Date
			record.Date = time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
		}
		if row.ProjectID.Valid {
			record.ProjectID = row.ProjectID.StringVal
		}
		if row.ProjectName.Valid {
			record.ProjectName = row.ProjectName.StringVal
		}
		if row.Service.Valid {
			record.Service = row.Service.StringVal
		}
		if row.SKU.Valid {
			record.SKU = row.SKU.StringVal
		}
		if row.Region.Valid {
			record.Region = row.Region.StringVal
		}
		if row.UsageAmount.Valid {
			record.UsageAmount = row.UsageAmount.Float64
		}
		if row.UsageUnit.Valid {
			record.UsageUnit = row.UsageUnit.StringVal
		}
		if row.Cost.Valid {
			record.Cost = row.Cost.Float64
		}

		// Estimate emissions if not available from Carbon Footprint
		record.EmissionsKgCO2 = c.estimateEmissions(record.Service, record.Region, record.UsageAmount, record.UsageUnit)

		records = append(records, record)
	}

	return records, nil
}

func (c *GCPConnector) estimateEmissions(service, region string, usage float64, unit string) float64 {
	// Convert usage to estimated kWh based on service type
	var estimatedKWh float64

	switch {
	case contains(service, "Compute Engine"):
		// 1 vCPU-hour ≈ 0.02 kWh
		if unit == "core-hour" || unit == "seconds" {
			if unit == "seconds" {
				usage = usage / 3600 // Convert to hours
			}
			estimatedKWh = usage * 0.02
		}
	case contains(service, "Cloud Storage"):
		// 1 GB-month ≈ 0.00001 kWh
		if unit == "byte-seconds" {
			usage = usage / (1e9 * 2592000) // Convert to GB-months
		}
		estimatedKWh = usage * 0.00001
	case contains(service, "BigQuery"):
		// 1 slot-hour ≈ 0.015 kWh
		if unit == "slot-hour" {
			estimatedKWh = usage * 0.015
		}
	}

	// Apply regional emission factor
	factor, ok := c.config.EmissionFactors[region]
	if !ok {
		factor = 0.4 // Default
	}

	return estimatedKWh * factor
}

func defaultGCPEmissionFactors() map[string]float64 {
	return map[string]float64{
		"us-central1":          0.4790, // Iowa
		"us-east1":             0.3854, // South Carolina
		"us-east4":             0.3854, // Virginia
		"us-west1":             0.2451, // Oregon
		"us-west2":             0.2451, // California
		"europe-west1":         0.3380, // Belgium
		"europe-west2":         0.2331, // London
		"europe-west3":         0.3380, // Frankfurt
		"europe-west4":         0.3380, // Netherlands
		"asia-east1":           0.5540, // Taiwan
		"asia-northeast1":      0.4630, // Tokyo
		"asia-southeast1":      0.4990, // Singapore
		"australia-southeast1": 0.7900, // Sydney
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

// Close closes the BigQuery client
func (c *GCPConnector) Close() error {
	return c.client.Close()
}
