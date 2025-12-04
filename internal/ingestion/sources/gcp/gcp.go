// Package gcp provides an ingestion adapter for Google Cloud Platform carbon emissions data.
// It fetches emissions data from GCP Carbon Footprint reports exported to BigQuery
// and the Cloud Billing API.
package gcp

import (
	"cloud.google.com/go/bigquery"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/example/offgridflow/internal/ingestion"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// =============================================================================
// Configuration
// =============================================================================

// Config holds GCP adapter configuration.
type Config struct {
	// ProjectID is the GCP project ID.
	ProjectID string `json:"project_id"`

	// ServiceAccountKey is the JSON service account key (for non-GCE environments).
	ServiceAccountKey string `json:"-"` // Excluded from JSON

	// BillingAccountID is the GCP billing account ID.
	BillingAccountID string `json:"billing_account_id"`

	// BigQueryDataset is the dataset containing carbon footprint exports.
	BigQueryDataset string `json:"bigquery_dataset,omitempty"`

	// BigQueryTable is the table containing carbon footprint data.
	BigQueryTable string `json:"bigquery_table,omitempty"`

	// OrgID is the OffGridFlow organization ID to associate activities with.
	OrgID string `json:"org_id"`

	// StartDate is the beginning of the date range to fetch (inclusive).
	StartDate time.Time `json:"start_date"`

	// EndDate is the end of the date range to fetch (exclusive).
	EndDate time.Time `json:"end_date"`

	// HTTPClient allows injecting a custom HTTP client for testing.
	HTTPClient *http.Client `json:"-"`
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	if c.ProjectID == "" {
		return fmt.Errorf("gcp: project_id is required")
	}
	if c.OrgID == "" {
		return fmt.Errorf("gcp: org_id is required")
	}
	return nil
}

// =============================================================================
// GCP API Response Types
// =============================================================================

// CarbonFootprintReport represents the GCP Carbon Footprint export structure.
type CarbonFootprintReport struct {
	Records []CarbonRecord `json:"records"`
}

// CarbonRecord represents a single carbon emissions record from GCP.
type CarbonRecord struct {
	BillingAccountID     string   `json:"billing_account_id"`
	Project              Project  `json:"project"`
	Service              Service  `json:"service"`
	Location             Location `json:"location"`
	UsageMonth           string   `json:"usage_month"` // Format: YYYYMM
	CarbonFootprintKgCO2 float64  `json:"carbon_footprint_kg_co2"`
	CarbonModelVersion   string   `json:"carbon_model_version"`
	Scope1Emissions      float64  `json:"scope_1_emissions_kg_co2,omitempty"`
	Scope2Emissions      float64  `json:"scope_2_emissions_kg_co2,omitempty"`
	Scope3Emissions      float64  `json:"scope_3_emissions_kg_co2,omitempty"`
	ElectricityKWh       float64  `json:"electricity_consumption_kwh,omitempty"`
	CFEScore             float64  `json:"carbon_free_energy_score,omitempty"` // Percentage 0-100
}

// Project represents a GCP project.
type Project struct {
	ID     string `json:"id"`
	Number string `json:"number"`
	Name   string `json:"name"`
}

// Service represents a GCP service.
type Service struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// Location represents a GCP location.
type Location struct {
	Location string `json:"location"` // e.g., "us-central1"
	Country  string `json:"country"`
	Region   string `json:"region"`
}

// BigQueryResponse represents a BigQuery query result.
type BigQueryResponse struct {
	Kind                string        `json:"kind"`
	Schema              Schema        `json:"schema"`
	JobReference        JobReference  `json:"jobReference"`
	TotalRows           string        `json:"totalRows"`
	Rows                []BigQueryRow `json:"rows"`
	TotalBytesProcessed string        `json:"totalBytesProcessed"`
	JobComplete         bool          `json:"jobComplete"`
	PageToken           string        `json:"pageToken,omitempty"`
}

// Schema represents the BigQuery schema.
type Schema struct {
	Fields []Field `json:"fields"`
}

// Field represents a BigQuery schema field.
type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Mode string `json:"mode"`
}

// JobReference identifies a BigQuery job.
type JobReference struct {
	ProjectID string `json:"projectId"`
	JobID     string `json:"jobId"`
	Location  string `json:"location"`
}

// BigQueryRow represents a row in BigQuery results.
type BigQueryRow struct {
	F []BigQueryValue `json:"f"`
}

// BigQueryValue represents a value in a BigQuery row.
type BigQueryValue struct {
	V interface{} `json:"v"`
}

// OAuthTokenResponse represents the Google OAuth token response.
type OAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// =============================================================================
// Adapter Implementation
// =============================================================================

// Adapter ingests carbon emissions data from Google Cloud Platform.
type Adapter struct {
	config      Config
	client      *http.Client
	accessToken string
	tokenExpiry time.Time
}

// NewAdapter creates a new GCP ingestion adapter.
func NewAdapter(cfg Config) (*Adapter, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{
			Timeout: 60 * time.Second, // BigQuery queries can be slow
		}
	}

	return &Adapter{
		config: cfg,
		client: client,
	}, nil
}

// Ingest fetches carbon emissions data from GCP and returns activities.
func (a *Adapter) Ingest(ctx context.Context) ([]ingestion.Activity, error) {
	var records []CarbonRecord
	var err error

	// Try BigQuery first if configured
	err = ingestion.WithRetry(ctx, 3, 2*time.Second, func() error {
		if a.config.BigQueryDataset != "" && a.config.BigQueryTable != "" {
			records, err = a.fetchFromBigQueryNative(ctx)
			return err
		}
		records, err = a.fetchFromCarbonAPI(ctx)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("gcp: failed to fetch carbon data: %w", err)
	}

	// Convert to activities
	activities := a.convertToActivities(records)

	return activities, nil
}

// ensureAuthenticated obtains or refreshes the Google OAuth access token.
func (a *Adapter) ensureAuthenticated(ctx context.Context) error {
	// Check if we have a valid token
	if a.accessToken != "" && time.Now().Before(a.tokenExpiry) {
		return nil
	}

	// If running on GCE, use metadata server
	// Otherwise, use service account key
	if a.config.ServiceAccountKey == "" {
		return a.authenticateWithMetadata(ctx)
	}
	return a.authenticateWithServiceAccount(ctx)
}

// authenticateWithMetadata uses the GCE metadata server.
func (a *Adapter) authenticateWithMetadata(ctx context.Context) error {
	endpoint := "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Metadata-Flavor", "Google")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("metadata request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("metadata request failed with status %d", resp.StatusCode)
	}

	var tokenResp OAuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	a.accessToken = tokenResp.AccessToken
	a.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	return nil
}

// authenticateWithServiceAccount uses a service account key for authentication.
func (a *Adapter) authenticateWithServiceAccount(ctx context.Context) error {
	// Note: In production, use google.golang.org/api/option with credentials
	// This is a simplified implementation
	endpoint := "https://oauth2.googleapis.com/token"

	// Parse the service account key
	var saKey struct {
		ClientEmail string `json:"client_email"`
		PrivateKey  string `json:"private_key"`
		TokenURI    string `json:"token_uri"`
	}
	if err := json.Unmarshal([]byte(a.config.ServiceAccountKey), &saKey); err != nil {
		return fmt.Errorf("invalid service account key: %w", err)
	}

	// Create signed JWT for service account authentication
	// Reference: https://cloud.google.com/iam/docs/creating-managing-service-account-keys
	now := time.Now()
	claims := map[string]interface{}{
		"iss":   saKey.ClientEmail,
		"scope": "https://www.googleapis.com/auth/cloud-platform",
		"aud":   endpoint,
		"exp":   now.Add(1 * time.Hour).Unix(),
		"iat":   now.Unix(),
	}

	// For production: implement JWT signing with crypto/rsa
	// The service account private key would be used to sign the JWT
	// Then exchange the signed JWT for an OAuth2 access token
	//
	// Example flow:
	// 1. Sign JWT with private key from saKey.PrivateKey
	// 2. POST signed JWT to endpoint
	// 3. Receive access token in response
	// 4. Use access token for API calls
	//
	// For now, recommend using the official google-cloud-go SDK which handles this automatically
	_ = claims
	return fmt.Errorf("gcp: service account authentication requires google-cloud-go SDK for production use - install: go get cloud.google.com/go/bigquery")
}

// fetchFromBigQuery queries carbon footprint data from BigQuery export.
// This function provides guidance for production BigQuery integration.
func (a *Adapter) fetchFromBigQuery(ctx context.Context) ([]CarbonRecord, error) {
	// For production BigQuery integration:
	// 1. Install google-cloud-go SDK: go get cloud.google.com/go/bigquery
	// 2. Import: "cloud.google.com/go/bigquery"
	// 3. Create BigQuery client with service account credentials
	// 4. Query the carbon footprint dataset
	//
	// Example production code:
	// 
	// client, err := bigquery.NewClient(ctx, a.config.ProjectID, option.WithCredentialsFile(a.config.ServiceAccountKey))
	// if err != nil {
	//     return nil, fmt.Errorf("gcp: failed to create BigQuery client: %w", err)
	// }
	// defer client.Close()
	//
	// query := client.Query(`
	//     SELECT billing_account_id, project_id, service_id, region, usage_month,
	//            carbon_footprint_kg_co2, scope_1_emissions_kg_co2, scope_2_emissions_kg_co2
	//     FROM ` + "`" + a.config.BigQueryDataset + "." + a.config.BigQueryTable + "`" + `
	//     WHERE usage_month >= @start_month AND usage_month < @end_month
	// `)
	// query.Parameters = []bigquery.QueryParameter{
	//     {Name: "start_month", Value: a.config.StartDate.Format("200601")},
	//     {Name: "end_month", Value: a.config.EndDate.Format("200601")},
	// }
	//
	// it, err := query.Read(ctx)
	// ... process results ...
	
	return nil, fmt.Errorf("gcp: BigQuery integration requires google-cloud-go SDK - run: go get cloud.google.com/go/bigquery")
}

// fetchFromCarbonAPI fetches carbon footprint data from the API.
// GCP provides carbon footprint data through BigQuery exports rather than a direct API.
// This method returns an empty slice and directs users to use BigQuery integration.
func (a *Adapter) fetchFromCarbonAPI(ctx context.Context) ([]CarbonRecord, error) {
	// GCP's recommended approach for carbon footprint data:
	// 1. Enable Carbon Footprint export in Cloud Console
	// 2. Export data to BigQuery dataset
	// 3. Use fetchFromBigQuery() method instead
	//
	// Reference: https://cloud.google.com/carbon-footprint
	//
	// If you need API-based access, use the BigQuery API with the carbon footprint dataset
	// Example dataset: `project.region.INFORMATION_SCHEMA.CARBON_FOOTPRINT_BY_PROJECT`
	//
	// Note: This is informational only - use BigQuery integration for production
	
	return []CarbonRecord{}, nil
}

// fetchFromBigQueryNative queries carbon footprint data using the native BigQuery client
func (a *Adapter) fetchFromBigQueryNative(ctx context.Context) ([]CarbonRecord, error) {
	var opts []option.ClientOption
	
	if a.config.ServiceAccountKey != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(a.config.ServiceAccountKey)))
	}

	client, err := bigquery.NewClient(ctx, a.config.ProjectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("gcp: failed to create BigQuery client: %w", err)
	}
	defer client.Close()

	// Build the query
	query := client.Query(fmt.Sprintf(`
		SELECT 
			billing_account_id,
			project.id as project_id,
			project.number as project_number,
			project.name as project_name,
			service.id as service_id,
			service.description as service_description,
			location.location as region,
			location.country as country,
			usage_month,
			carbon_footprint_kg_co2,
			carbon_model_version,
			IFNULL(carbon_footprint_breakdown.scope1, 0) as scope_1_emissions_kg_co2,
			IFNULL(carbon_footprint_breakdown.scope2, 0) as scope_2_emissions_kg_co2,
			IFNULL(carbon_footprint_breakdown.scope3, 0) as scope_3_emissions_kg_co2,
			IFNULL(carbon_offsets_kg_co2, 0) as carbon_offsets_kg_co2
		FROM %s.%s
		WHERE usage_month >= @start_month
		AND usage_month < @end_month
		ORDER BY usage_month, project_id, service_id
	`, a.config.BigQueryDataset, a.config.BigQueryTable))

	query.Parameters = []bigquery.QueryParameter{
		{Name: "start_month", Value: a.config.StartDate.Format("200601")},
		{Name: "end_month", Value: a.config.EndDate.Format("200601")},
	}

	it, err := query.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("gcp: failed to execute BigQuery query: %w", err)
	}

	var records []CarbonRecord
	for {
		var row struct {
			BillingAccountID          string  `bigquery:"billing_account_id"`
			ProjectID                 string  `bigquery:"project_id"`
			ProjectNumber             string  `bigquery:"project_number"`
			ProjectName               string  `bigquery:"project_name"`
			ServiceID                 string  `bigquery:"service_id"`
			ServiceDescription        string  `bigquery:"service_description"`
			Region                    string  `bigquery:"region"`
			Country                   string  `bigquery:"country"`
			UsageMonth                string  `bigquery:"usage_month"`
			CarbonFootprintKgCO2      float64 `bigquery:"carbon_footprint_kg_co2"`
			CarbonModelVersion        string  `bigquery:"carbon_model_version"`
			Scope1Emissions           float64 `bigquery:"scope_1_emissions_kg_co2"`
			Scope2Emissions           float64 `bigquery:"scope_2_emissions_kg_co2"`
			Scope3Emissions           float64 `bigquery:"scope_3_emissions_kg_co2"`
			ElectricityConsumptionKwh float64 `bigquery:"electricity_consumption_kwh"`
			CarbonFreeEnergyPct       float64 `bigquery:"carbon_free_energy_percentage"`
		}

		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("gcp: iterate bigquery results: %w", err)
		}

		records = append(records, CarbonRecord{
			BillingAccountID: row.BillingAccountID,
			Project: Project{
				ID:     row.ProjectID,
				Number: row.ProjectNumber,
				Name:   row.ProjectName,
			},
			Service: Service{
				ID:          row.ServiceID,
				Description: row.ServiceDescription,
			},
			Location: Location{
				Location: row.Region,
				Country:  row.Country,
			},
			UsageMonth:           row.UsageMonth,
			CarbonFootprintKgCO2: row.CarbonFootprintKgCO2,
			CarbonModelVersion:   row.CarbonModelVersion,
			Scope1Emissions:      row.Scope1Emissions,
			Scope2Emissions:      row.Scope2Emissions,
			Scope3Emissions:      row.Scope3Emissions,
			ElectricityKWh:       row.ElectricityConsumptionKwh,
			CFEScore:             row.CarbonFreeEnergyPct,
		})
	}

	return records, nil
}

// parseBigQueryResults converts BigQuery rows to CarbonRecords.
func (a *Adapter) parseBigQueryResults(result BigQueryResponse) []CarbonRecord {
	records := make([]CarbonRecord, 0, len(result.Rows))

	for _, row := range result.Rows {
		if len(row.F) < 14 {
			continue // Skip malformed rows
		}

		record := CarbonRecord{
			BillingAccountID: toString(row.F[0].V),
			Project: Project{
				ID:   toString(row.F[1].V),
				Name: toString(row.F[2].V),
			},
			Service: Service{
				ID:          toString(row.F[3].V),
				Description: toString(row.F[4].V),
			},
			Location: Location{
				Location: toString(row.F[5].V),
				Country:  toString(row.F[6].V),
			},
			UsageMonth:           toString(row.F[7].V),
			Scope1Emissions:      toFloat(row.F[8].V),
			Scope2Emissions:      toFloat(row.F[9].V),
			Scope3Emissions:      toFloat(row.F[10].V),
			CarbonFootprintKgCO2: toFloat(row.F[11].V),
			CFEScore:             toFloat(row.F[13].V),
		}

		records = append(records, record)
	}

	return records
}

// convertToActivities transforms GCP carbon data into OffGridFlow activities.
func (a *Adapter) convertToActivities(records []CarbonRecord) []ingestion.Activity {
	activities := make([]ingestion.Activity, 0)
	now := time.Now().UTC()

	for _, record := range records {
		// Parse the usage month
		periodStart, err := time.Parse("200601", record.UsageMonth)
		if err != nil {
			continue
		}
		periodEnd := periodStart.AddDate(0, 1, 0)

		// Map GCP location to OffGridFlow location code
		location := mapGCPRegion(record.Location.Location)

		// Create an activity for total emissions
		totalEmissions := record.CarbonFootprintKgCO2 / 1000 // Convert kg to tonnes

		if totalEmissions > 0 {
			activity := ingestion.Activity{
				ID:          uuid.NewString(),
				Source:      "gcp_carbon_footprint",
				Category:    categorizeGCPService(record.Service.ID),
				Location:    location,
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
				Quantity:    totalEmissions,
				Unit:        "tonne",
				OrgID:       a.config.OrgID,
				Metadata: map[string]string{
					"gcp_project_id":       record.Project.ID,
					"gcp_project_name":     record.Project.Name,
					"gcp_service_id":       record.Service.ID,
					"gcp_service_name":     record.Service.Description,
					"gcp_region":           record.Location.Location,
					"gcp_country":          record.Location.Country,
					"usage_month":          record.UsageMonth,
					"cfe_score":            fmt.Sprintf("%.1f", record.CFEScore),
					"scope1_kg":            fmt.Sprintf("%.2f", record.Scope1Emissions),
					"scope2_kg":            fmt.Sprintf("%.2f", record.Scope2Emissions),
					"scope3_kg":            fmt.Sprintf("%.2f", record.Scope3Emissions),
					"carbon_model_version": record.CarbonModelVersion,
					"data_source":          "gcp_carbon_footprint_export",
				},
				CreatedAt:   now,
				DataQuality: "measured",
				ExternalID:  fmt.Sprintf("gcp_%s_%s_%s_%s", record.Project.ID, record.Service.ID, record.Location.Location, record.UsageMonth),
			}

			activities = append(activities, activity)
		}

		// Also create granular scope activities if available
		if record.Scope2Emissions > 0 {
			activities = append(activities, ingestion.Activity{
				ID:          uuid.NewString(),
				Source:      "gcp_carbon_footprint",
				Category:    "cloud_compute_scope2",
				Location:    location,
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
				Quantity:    record.Scope2Emissions / 1000,
				Unit:        "tonne",
				OrgID:       a.config.OrgID,
				Metadata: map[string]string{
					"gcp_project_id":  record.Project.ID,
					"gcp_service_id":  record.Service.ID,
					"gcp_region":      record.Location.Location,
					"emission_scope":  "scope2",
					"electricity_kwh": fmt.Sprintf("%.2f", record.ElectricityKWh),
					"cfe_score":       fmt.Sprintf("%.1f", record.CFEScore),
					"data_source":     "gcp_carbon_footprint_export",
				},
				CreatedAt:   now,
				DataQuality: "measured",
				ExternalID:  fmt.Sprintf("gcp_%s_%s_%s_%s_scope2", record.Project.ID, record.Service.ID, record.Location.Location, record.UsageMonth),
			})
		}
	}

	return activities
}

// mapGCPRegion converts GCP region codes to OffGridFlow location codes.
func mapGCPRegion(gcpRegion string) string {
	regionMap := map[string]string{
		"us-central1":             "US-CENTRAL",
		"us-east1":                "US-EAST",
		"us-east4":                "US-EAST",
		"us-east5":                "US-EAST",
		"us-west1":                "US-WEST",
		"us-west2":                "US-WEST",
		"us-west3":                "US-WEST",
		"us-west4":                "US-WEST",
		"us-south1":               "US-CENTRAL",
		"northamerica-northeast1": "US-EAST",
		"northamerica-northeast2": "US-EAST",
		"southamerica-east1":      "LATAM",
		"southamerica-west1":      "LATAM",
		"europe-west1":            "EU-WEST",
		"europe-west2":            "EU-WEST",
		"europe-west3":            "EU-CENTRAL",
		"europe-west4":            "EU-WEST",
		"europe-west6":            "EU-CENTRAL",
		"europe-west8":            "EU-WEST",
		"europe-west9":            "EU-WEST",
		"europe-north1":           "EU-NORTH",
		"europe-central2":         "EU-CENTRAL",
		"asia-east1":              "ASIA-PACIFIC",
		"asia-east2":              "ASIA-PACIFIC",
		"asia-northeast1":         "ASIA-PACIFIC",
		"asia-northeast2":         "ASIA-PACIFIC",
		"asia-northeast3":         "ASIA-PACIFIC",
		"asia-south1":             "ASIA-PACIFIC",
		"asia-south2":             "ASIA-PACIFIC",
		"asia-southeast1":         "ASIA-PACIFIC",
		"asia-southeast2":         "ASIA-PACIFIC",
		"australia-southeast1":    "ASIA-PACIFIC",
		"australia-southeast2":    "ASIA-PACIFIC",
		"me-west1":                "MENA",
		"me-central1":             "MENA",
		"africa-south1":           "AFRICA",
	}

	if location, ok := regionMap[strings.ToLower(gcpRegion)]; ok {
		return location
	}
	return "GLOBAL"
}

// categorizeGCPService maps GCP service IDs to emission categories.
func categorizeGCPService(serviceID string) string {
	categoryMap := map[string]string{
		"6F81-5844-456A": "cloud_compute",            // Compute Engine
		"95FF-2EF5-5EA1": "cloud_database",           // Cloud SQL
		"152E-C115-5142": "cloud_storage",            // Cloud Storage
		"29E7-DA93-CA13": "cloud_compute_serverless", // Cloud Functions
		"2062-016F-44A2": "cloud_compute_kubernetes", // GKE
		"24E6-581D-38E5": "cloud_compute_serverless", // Cloud Run
		"D97E-AB26-5D95": "cloud_data",               // BigQuery
		"6F52-F27E-4C48": "cloud_network",            // Cloud CDN
		"E505-1604-58F8": "cloud_database",           // Cloud Spanner
	}

	if category, ok := categoryMap[serviceID]; ok {
		return category
	}
	return "cloud_other"
}

// =============================================================================
// Helper Functions
// =============================================================================

// toString safely converts an interface{} to string.
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// toFloat safely converts an interface{} to float64.
func toFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	case int:
		return float64(val)
	case int64:
		return float64(val)
	}
	return 0
}
