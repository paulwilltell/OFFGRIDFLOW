package emissions

import (
	"testing"
)

func TestScope_String(t *testing.T) {
	tests := []struct {
		scope    Scope
		expected string
	}{
		{Scope1, "Scope 1"},
		{Scope2, "Scope 2"},
		{Scope3, "Scope 3"},
		{ScopeUnspecified, "Unknown"},
		{Scope(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.scope.String()
			if result != tt.expected {
				t.Errorf("Scope.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestScope_IsValid(t *testing.T) {
	tests := []struct {
		scope    Scope
		expected bool
	}{
		{Scope1, true},
		{Scope2, true},
		{Scope3, true},
		{ScopeUnspecified, false},
		{Scope(0), false},
		{Scope(99), false},
	}

	for _, tt := range tests {
		t.Run(tt.scope.String(), func(t *testing.T) {
			result := tt.scope.IsValid()
			if result != tt.expected {
				t.Errorf("Scope.IsValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}
