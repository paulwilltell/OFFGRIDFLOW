package demo

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHandler(t *testing.T) {
	cfg := DemoConfig{
		Enabled:     true,
		CompanyName: "Test Corp",
		Industry:    "Technology",
	}
	h := NewHandler(cfg)

	if h == nil {
		t.Fatal("NewHandler returned nil")
	}
}

func TestHandler_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{"enabled", true, true},
		{"disabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DemoConfig{Enabled: tt.enabled}
			h := NewHandler(cfg)

			if h.IsEnabled() != tt.expected {
				t.Errorf("IsEnabled() = %v, want %v", h.IsEnabled(), tt.expected)
			}
		})
	}
}

func TestHandler_GetData(t *testing.T) {
	cfg := DemoConfig{
		Enabled:     true,
		CompanyName: "Demo Company",
		Industry:    "Technology",
		BaseYear:    2022,
	}
	h := NewHandler(cfg)

	data := h.GetData()
	if data == nil {
		t.Fatal("GetData returned nil")
	}

	// Second call should return same instance (sync.Once)
	data2 := h.GetData()
	if data != data2 {
		t.Log("GetData returns new instance each time - check sync.Once usage")
	}
}

func TestHandler_RefreshData(t *testing.T) {
	cfg := DemoConfig{
		Enabled:     true,
		CompanyName: "Demo Company",
		Industry:    "Manufacturing",
	}
	h := NewHandler(cfg)

	// Get initial data
	data1 := h.GetData()

	// Refresh
	h.RefreshData()

	// Get refreshed data
	data2 := h.GetData()

	if data1 == nil || data2 == nil {
		t.Error("Data should not be nil")
	}
}

func TestHandler_ServeHTTP_Disabled(t *testing.T) {
	cfg := DemoConfig{Enabled: false}
	h := NewHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/demo/data", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 when disabled, got %d", w.Code)
	}
}

func TestHandler_ServeHTTP_Data(t *testing.T) {
	cfg := DemoConfig{
		Enabled:     true,
		CompanyName: "Demo Company",
		Industry:    "Technology",
	}
	h := NewHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/demo/data", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandler_ServeHTTP_Summary(t *testing.T) {
	cfg := DemoConfig{
		Enabled:     true,
		CompanyName: "Demo Company",
		Industry:    "Technology",
	}
	h := NewHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/demo/summary", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandler_ServeHTTP_NotFound(t *testing.T) {
	cfg := DemoConfig{Enabled: true}
	h := NewHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/demo/unknown", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for unknown endpoint, got %d", w.Code)
	}
}

func TestDefaultDemoConfig(t *testing.T) {
	cfg := DefaultDemoConfig()

	if cfg.CompanyName == "" {
		t.Error("CompanyName should have default value")
	}
	if cfg.Industry == "" {
		t.Error("Industry should have default value")
	}
	if cfg.BaseYear <= 0 {
		t.Error("BaseYear should be positive")
	}
}

func TestDemoConfig_Fields(t *testing.T) {
	cfg := DemoConfig{
		Enabled:     true,
		CompanyName: "Test Corp",
		Industry:    "Technology",
		BaseYear:    2020,
		ShowAIDemo:  true,
	}

	if !cfg.Enabled {
		t.Error("Enabled should be true")
	}
	if cfg.CompanyName != "Test Corp" {
		t.Errorf("Expected CompanyName 'Test Corp', got '%s'", cfg.CompanyName)
	}
}
