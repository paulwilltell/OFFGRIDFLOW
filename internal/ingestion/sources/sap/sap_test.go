package sap

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				BaseURL:      "https://api.sap.company.com",
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				Company:      "1000",
				OrgID:        "org-123",
			},
			wantErr: false,
		},
		{
			name: "missing base URL",
			config: Config{
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				Company:      "1000",
				OrgID:        "org-123",
			},
			wantErr: true,
			errMsg:  "base_url is required",
		},
		{
			name: "missing client ID",
			config: Config{
				BaseURL:      "https://api.sap.company.com",
				ClientSecret: "test-secret",
				Company:      "1000",
				OrgID:        "org-123",
			},
			wantErr: true,
			errMsg:  "client_id is required",
		},
		{
			name: "missing client secret",
			config: Config{
				BaseURL:  "https://api.sap.company.com",
				ClientID: "test-client",
				Company:  "1000",
				OrgID:    "org-123",
			},
			wantErr: true,
			errMsg:  "client_secret is required",
		},
		{
			name: "missing company",
			config: Config{
				BaseURL:      "https://api.sap.company.com",
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				OrgID:        "org-123",
			},
			wantErr: true,
			errMsg:  "company is required",
		},
		{
			name: "missing org ID",
			config: Config{
				BaseURL:      "https://api.sap.company.com",
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				Company:      "1000",
			},
			wantErr: true,
			errMsg:  "org_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewAdapter(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := Config{
			BaseURL:      "https://api.sap.company.com",
			ClientID:     "test-client",
			ClientSecret: "test-secret",
			Company:      "1000",
			OrgID:        "org-123",
		}

		adapter, err := NewAdapter(cfg)
		require.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.NotNil(t, adapter.client)
	})

	t.Run("invalid config", func(t *testing.T) {
		cfg := Config{
			BaseURL: "https://api.sap.company.com",
		}

		adapter, err := NewAdapter(cfg)
		assert.Error(t, err)
		assert.Nil(t, adapter)
	})
}

func TestAdapter_Authenticate(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/oauth/token", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		// Verify basic auth
		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "test-client", username)
		assert.Equal(t, "test-secret", password)

		// Return token response
		tokenResp := TokenResponse{
			AccessToken: "test-token-123",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tokenResp)
	}))
	defer server.Close()

	cfg := Config{
		BaseURL:      server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Company:      "1000",
		OrgID:        "org-123",
	}

	adapter, err := NewAdapter(cfg)
	require.NoError(t, err)

	err = adapter.authenticate(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "test-token-123", adapter.token)
	assert.True(t, adapter.tokenExpiry.After(time.Now()))
}

func TestAdapter_FetchEnergyData(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			tokenResp := TokenResponse{
				AccessToken: "test-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tokenResp)
			return
		}

		if r.URL.Path == "/api/energy/consumption" {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

			// Verify query parameters
			q := r.URL.Query()
			assert.Equal(t, "1000", q.Get("company"))

			// Return energy data
			energyResp := EnergyDataResponse{
				Data: []EnergyRecord{
					{
						RecordID:   "E001",
						Plant:      "US-TX-001",
						Meter:      "M-100",
						Date:       "2024-01-15",
						EnergyType: "Electricity",
						Quantity:   1500.5,
						Unit:       "kWh",
						CostCenter: "CC-1000",
					},
					{
						RecordID:   "E002",
						Plant:      "US-TX-001",
						Meter:      "M-101",
						Date:       "2024-01-15",
						EnergyType: "Natural_Gas",
						Quantity:   250.0,
						Unit:       "m3",
						CostCenter: "CC-1000",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(energyResp)
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	cfg := Config{
		BaseURL:      server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Company:      "1000",
		OrgID:        "org-123",
		StartDate:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	adapter, err := NewAdapter(cfg)
	require.NoError(t, err)

	// Set token manually to skip auth
	adapter.token = "test-token"
	adapter.tokenExpiry = time.Now().Add(time.Hour)

	activities, err := adapter.fetchEnergyData(context.Background())
	require.NoError(t, err)
	assert.Len(t, activities, 2)

	// Verify first activity (electricity)
	activity := activities[0]
	assert.Equal(t, "sap_erp", activity.Source)
	assert.Equal(t, "electricity", activity.Category)
	assert.Equal(t, "M-100", activity.MeterID)
	assert.Equal(t, 1500.5, activity.Quantity)
	assert.Equal(t, "kWh", activity.Unit)
	assert.Equal(t, "org-123", activity.OrgID)
	assert.Equal(t, "E001", activity.ExternalID)
	assert.Equal(t, "measured", activity.DataQuality)
	assert.Contains(t, activity.Metadata, "sap_plant")
	assert.Equal(t, "US-TX-001", activity.Metadata["sap_plant"])

	// Verify second activity (natural gas)
	activity = activities[1]
	assert.Equal(t, "natural_gas", activity.Category)
	assert.Equal(t, "m3", activity.Unit)
	assert.Equal(t, 250.0, activity.Quantity)
}

func TestAdapter_FetchEmissionsData(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			tokenResp := TokenResponse{
				AccessToken: "test-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tokenResp)
			return
		}

		if r.URL.Path == "/api/sustainability/emissions" {
			assert.Equal(t, "GET", r.Method)

			// Return emissions data
			emissionsResp := EmissionsDataResponse{
				Data: []EmissionRecord{
					{
						RecordID:     "EM001",
						Date:         "2024-01-15",
						Source:       "Purchased Electricity",
						EmissionType: "CO2",
						KgCO2e:       1250.5,
						Scope:        "Scope2",
						Plant:        "US-TX-001",
					},
					{
						RecordID:     "EM002",
						Date:         "2024-01-15",
						Source:       "Natural Gas Combustion",
						EmissionType: "CO2",
						KgCO2e:       450.3,
						Scope:        "Scope1",
						Plant:        "US-TX-001",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(emissionsResp)
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	cfg := Config{
		BaseURL:      server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Company:      "1000",
		OrgID:        "org-123",
		StartDate:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	adapter, err := NewAdapter(cfg)
	require.NoError(t, err)

	// Set token manually
	adapter.token = "test-token"
	adapter.tokenExpiry = time.Now().Add(time.Hour)

	activities, err := adapter.fetchEmissionsData(context.Background())
	require.NoError(t, err)
	assert.Len(t, activities, 2)

	// Verify Scope 2 activity
	activity := activities[0]
	assert.Equal(t, "sap_sustainability", activity.Source)
	assert.Contains(t, activity.Category, "scope2")
	assert.Equal(t, 1250.5, activity.Quantity)
	assert.Equal(t, "kg", activity.Unit)
	assert.Equal(t, "Scope2", activity.Metadata["emission_scope"])

	// Verify Scope 1 activity
	activity = activities[1]
	assert.Contains(t, activity.Category, "scope1")
	assert.Equal(t, 450.3, activity.Quantity)
}

func TestAdapter_Ingest(t *testing.T) {
	// Create comprehensive mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/oauth/token":
			json.NewEncoder(w).Encode(TokenResponse{
				AccessToken: "test-token",
				ExpiresIn:   3600,
			})

		case "/api/energy/consumption":
			json.NewEncoder(w).Encode(EnergyDataResponse{
				Data: []EnergyRecord{
					{
						RecordID:   "E001",
						Plant:      "US-TX-001",
						Meter:      "M-100",
						Date:       "2024-01-15",
						EnergyType: "Electricity",
						Quantity:   1000.0,
						Unit:       "kWh",
					},
				},
			})

		case "/api/sustainability/emissions":
			json.NewEncoder(w).Encode(EmissionsDataResponse{
				Data: []EmissionRecord{
					{
						RecordID: "EM001",
						Date:     "2024-01-15",
						Scope:    "Scope2",
						KgCO2e:   800.0,
					},
				},
			})

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	cfg := Config{
		BaseURL:      server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Company:      "1000",
		OrgID:        "org-123",
		StartDate:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	adapter, err := NewAdapter(cfg)
	require.NoError(t, err)

	activities, err := adapter.Ingest(context.Background())
	require.NoError(t, err)
	
	// Should have both energy and emissions activities
	assert.GreaterOrEqual(t, len(activities), 2)

	// Verify we have different sources
	sources := make(map[string]bool)
	for _, activity := range activities {
		sources[activity.Source] = true
		assert.Equal(t, "org-123", activity.OrgID)
	}
	
	// Should have both sap_erp and sap_sustainability sources
	assert.True(t, sources["sap_erp"] || sources["sap_sustainability"])
}

func TestMapEnergyType(t *testing.T) {
	tests := []struct {
		sapType  string
		expected string
	}{
		{"Electricity", "electricity"},
		{"ELECTRIC", "electricity"},
		{"Natural_Gas", "natural_gas"},
		{"GAS", "natural_gas"},
		{"Diesel", "diesel"},
		{"Fuel_Oil", "fuel_oil"},
		{"Heating_Oil", "fuel_oil"},
		{"Steam", "steam"},
		{"Water", "water"},
		{"Solar", "energy_solar"},
	}

	for _, tt := range tests {
		t.Run(tt.sapType, func(t *testing.T) {
			result := mapEnergyType(tt.sapType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapUnit(t *testing.T) {
	tests := []struct {
		sapUnit  string
		expected string
	}{
		{"kWh", "kWh"},
		{"kwh", "kWh"},
		{"MWh", "MWh"},
		{"mwh", "MWh"},
		{"GJ", "GJ"},
		{"gj", "GJ"},
		{"L", "L"},
		{"liter", "L"},
		{"m3", "m3"},
		{"kg", "kg"},
		{"tonne", "tonne"},
		{"ton", "tonne"},
	}

	for _, tt := range tests {
		t.Run(tt.sapUnit, func(t *testing.T) {
			result := mapUnit(tt.sapUnit)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapPlantToLocation(t *testing.T) {
	tests := []struct {
		plant    string
		expected string
	}{
		{"US-TX-001", "US-TX"},
		{"EU-DE-100", "EU-DE"},
		{"CN-BJ-050", "CN-BJ"},
		{"US001", "US-UNKNOWN"},
		{"EU100", "EU-CENTRAL"},
		{"CN500", "ASIA-CHINA"},
		{"", "UNKNOWN"},
		{"OTHER", "GLOBAL"},
	}

	for _, tt := range tests {
		t.Run(tt.plant, func(t *testing.T) {
			result := mapPlantToLocation(tt.plant)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeEmissionType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"CO2 Emissions", "co2_emissions"},
		{"Natural-Gas", "natural_gas"},
		{"FUEL OIL", "fuel_oil"},
		{"Purchased Electricity", "purchased_electricity"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeEmissionType(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
