package secrets

import (
	"context"
	"fmt"
	"os"
)

const (
	AWSAccessKeyID     = "OFFGRIDFLOW_AWS_ACCESS_KEY_ID"
	AWSSecretAccessKey = "OFFGRIDFLOW_AWS_SECRET_ACCESS_KEY"

	AzureClientSecret = "OFFGRIDFLOW_AZURE_CLIENT_SECRET"

	GCPServiceAccountKey = "OFFGRIDFLOW_GCP_SERVICE_ACCOUNT_KEY"

	SAPClientSecret = "OFFGRIDFLOW_SAP_CLIENT_SECRET"
)

// Provider resolves secrets by key names.
type Provider interface {
	Get(ctx context.Context, key string) (string, error)
}

// EnvProvider reads secrets from environment variables.
type EnvProvider struct{}

// NewEnvProvider returns a new EnvProvider instance.
func NewEnvProvider() *EnvProvider {
	return &EnvProvider{}
}

// Get returns the environment variable value for the provided key.
func (EnvProvider) Get(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("secret key missing")
	}
	if value := os.Getenv(key); value != "" {
		return value, nil
	}
	return "", fmt.Errorf("secret %s not found", key)
}

// Resolve prefers an explicit value, falling back to secret lookup when available.
func Resolve(ctx context.Context, provider Provider, explicit, key string) string {
	if explicit != "" {
		return explicit
	}
	if provider == nil {
		return ""
	}
	if val, err := provider.Get(ctx, key); err == nil {
		return val
	}
	return ""
}
