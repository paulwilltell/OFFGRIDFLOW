package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// =============================================================================
// Configuration Constants
// =============================================================================

const (
	// defaultOpenAIBaseURL is the standard OpenAI API endpoint.
	defaultOpenAIBaseURL = "https://api.openai.com/v1"

	// defaultOpenAIModel is the default model when none is specified.
	// gpt-4o-mini offers excellent price/performance for most use cases.
	defaultOpenAIModel = "gpt-4o-mini"

	// defaultOpenAITimeout is the maximum time for a single API request.
	// This should be generous enough for complex completions but not infinite.
	defaultOpenAITimeout = 60 * time.Second

	// defaultOpenAIMaxTokens is the default maximum tokens in the response.
	defaultOpenAIMaxTokens = 1024

	// defaultOpenAITemperature balances creativity and consistency.
	defaultOpenAITemperature = 0.7

	// maxResponseBodySize limits response body reads to prevent memory exhaustion.
	// 10MB should be more than sufficient for any chat completion response.
	maxResponseBodySize = 10 * 1024 * 1024

	// envKeyOpenAIAPIKey is the environment variable for the API key.
	envKeyOpenAIAPIKey = "OFFGRIDFLOW_OPENAI_API_KEY"

	// envKeyOpenAIModel is the environment variable for the model override.
	envKeyOpenAIModel = "OFFGRIDFLOW_OPENAI_MODEL"

	// envKeyOpenAIBaseURL is the environment variable for endpoint override.
	envKeyOpenAIBaseURL = "OFFGRIDFLOW_OPENAI_BASE_URL"
)

// defaultSystemPrompt provides domain-specific context for OffGridFlow.
const defaultSystemPrompt = `You are an AI assistant for OffGridFlow, a carbon accounting and emissions management platform.

Your expertise includes:
- Greenhouse gas (GHG) emissions calculations and reporting
- Scope 1, 2, and 3 emissions definitions and methodologies
- Emission factors and their regional variations
- Regulatory compliance frameworks: CSRD, SEC Climate Disclosure, CBAM, California regulations
- Sustainability strategies and carbon reduction pathways

Guidelines:
- Be concise, accurate, and cite specific regulations when relevant
- When explaining calculations, reference the GHG Protocol methodology
- Provide actionable recommendations grounded in data
- Acknowledge uncertainty when emission factors or data are estimates`

// =============================================================================
// Sentinel Errors
// =============================================================================

var (
	// ErrOpenAINotConfigured indicates the API key is missing or empty.
	ErrOpenAINotConfigured = errors.New("ai/openai: API key not configured")

	// ErrOpenAIRateLimited indicates the API returned a 429 status.
	ErrOpenAIRateLimited = errors.New("ai/openai: rate limited")

	// ErrOpenAIServerError indicates a 5xx response from OpenAI.
	ErrOpenAIServerError = errors.New("ai/openai: server error")

	// ErrOpenAIInvalidResponse indicates the response could not be parsed.
	ErrOpenAIInvalidResponse = errors.New("ai/openai: invalid response format")

	// ErrOpenAINoChoices indicates the response contained no completion choices.
	ErrOpenAINoChoices = errors.New("ai/openai: no choices in response")

	// ErrOpenAIContentFiltered indicates content was blocked by safety filters.
	ErrOpenAIContentFiltered = errors.New("ai/openai: content filtered")
)

// =============================================================================
// Provider Configuration
// =============================================================================

// OpenAIProviderConfig holds all configuration for the OpenAI provider.
// Use NewOpenAIProviderFromEnv for production or NewOpenAIProvider for tests.
type OpenAIProviderConfig struct {
	// APIKey is the OpenAI API key. Required.
	APIKey string

	// Model specifies which model to use (e.g., "gpt-4o", "gpt-4o-mini").
	// Defaults to defaultOpenAIModel if empty.
	Model string

	// BaseURL allows overriding the API endpoint for Azure OpenAI,
	// proxies, or compatible APIs. Defaults to defaultOpenAIBaseURL.
	BaseURL string

	// Timeout is the maximum duration for API requests.
	// Defaults to defaultOpenAITimeout if zero.
	Timeout time.Duration

	// DefaultSystem is the system prompt used when requests don't specify one.
	// Defaults to defaultSystemPrompt if empty.
	DefaultSystem string

	// DefaultMaxTokens is used when requests don't specify max_tokens.
	// Defaults to defaultOpenAIMaxTokens if zero.
	DefaultMaxTokens int

	// DefaultTemperature is used when requests don't specify temperature.
	// Defaults to defaultOpenAITemperature if zero.
	DefaultTemperature float64

	// HTTPClient allows injecting a custom HTTP client for testing.
	// If nil, a new client with the configured timeout is created.
	HTTPClient *http.Client
}

// validate checks the configuration and returns an error if invalid.
func (c *OpenAIProviderConfig) validate() error {
	if strings.TrimSpace(c.APIKey) == "" {
		return ErrOpenAINotConfigured
	}
	return nil
}

// applyDefaults fills in default values for unset fields.
func (c *OpenAIProviderConfig) applyDefaults() {
	c.APIKey = strings.TrimSpace(c.APIKey)
	c.BaseURL = strings.TrimSuffix(strings.TrimSpace(c.BaseURL), "/")
	c.Model = strings.TrimSpace(c.Model)
	c.DefaultSystem = strings.TrimSpace(c.DefaultSystem)

	if c.BaseURL == "" {
		c.BaseURL = defaultOpenAIBaseURL
	}
	if c.Model == "" {
		c.Model = defaultOpenAIModel
	}
	if c.Timeout <= 0 {
		c.Timeout = defaultOpenAITimeout
	}
	if c.DefaultSystem == "" {
		c.DefaultSystem = defaultSystemPrompt
	}
	if c.DefaultMaxTokens <= 0 {
		c.DefaultMaxTokens = defaultOpenAIMaxTokens
	}
	if c.DefaultTemperature <= 0 {
		c.DefaultTemperature = defaultOpenAITemperature
	}
}

// =============================================================================
// OpenAI Provider Implementation
// =============================================================================

// OpenAIProvider implements CloudProvider using the OpenAI Chat Completions API.
// It is safe for concurrent use.
type OpenAIProvider struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client

	defaultSystem      string
	defaultMaxTokens   int
	defaultTemperature float64

	// mu protects metrics fields
	mu           sync.RWMutex
	requestCount int64
	errorCount   int64
	totalLatency time.Duration
}

// NewOpenAIProvider creates a new OpenAI provider with the given configuration.
// Returns an error if the configuration is invalid.
func NewOpenAIProvider(cfg OpenAIProviderConfig) (*OpenAIProvider, error) {
	cfg.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: cfg.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		}
	}

	return &OpenAIProvider{
		apiKey:             cfg.APIKey,
		model:              cfg.Model,
		baseURL:            cfg.BaseURL,
		httpClient:         httpClient,
		defaultSystem:      cfg.DefaultSystem,
		defaultMaxTokens:   cfg.DefaultMaxTokens,
		defaultTemperature: cfg.DefaultTemperature,
	}, nil
}

// NewOpenAIProviderFromEnv creates a provider from environment variables.
//
// Environment variables:
//   - OFFGRIDFLOW_OPENAI_API_KEY (required): The OpenAI API key
//   - OFFGRIDFLOW_OPENAI_MODEL (optional): Model to use, defaults to gpt-4o-mini
//   - OFFGRIDFLOW_OPENAI_BASE_URL (optional): Override API endpoint
//
// Returns ErrOpenAINotConfigured if the API key is not set.
func NewOpenAIProviderFromEnv() (*OpenAIProvider, error) {
	apiKey := os.Getenv(envKeyOpenAIAPIKey)
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("%w: %s environment variable not set", ErrOpenAINotConfigured, envKeyOpenAIAPIKey)
	}

	return NewOpenAIProvider(OpenAIProviderConfig{
		APIKey:  apiKey,
		Model:   os.Getenv(envKeyOpenAIModel),
		BaseURL: os.Getenv(envKeyOpenAIBaseURL),
	})
}

// Chat sends a request to the OpenAI API and returns the response.
// It respects context cancellation and implements proper error handling.
func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	startTime := time.Now()

	// Validate request
	if err := req.Validate(); err != nil {
		return ChatResponse{}, err
	}

	// Check context before making request
	if err := ctx.Err(); err != nil {
		return ChatResponse{}, fmt.Errorf("ai/openai: %w", err)
	}

	// Build and execute request
	resp, err := p.doRequest(ctx, req)

	// Update metrics
	p.mu.Lock()
	p.requestCount++
	p.totalLatency += time.Since(startTime)
	if err != nil {
		p.errorCount++
	}
	p.mu.Unlock()

	if err != nil {
		return ChatResponse{}, err
	}

	resp.Latency = time.Since(startTime)
	return resp, nil
}

// doRequest performs the actual HTTP request to OpenAI.
func (p *OpenAIProvider) doRequest(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	// Build request payload
	payload := p.buildRequestPayload(req)

	body, err := json.Marshal(payload)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("ai/openai: failed to marshal request: %w", err)
	}

	// Create HTTP request
	endpoint := p.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("ai/openai: failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("User-Agent", "OffGridFlow/1.0")

	// Execute request
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		// Check if context was canceled
		if ctx.Err() != nil {
			return ChatResponse{}, fmt.Errorf("ai/openai: %w", ctx.Err())
		}
		return ChatResponse{}, fmt.Errorf("ai/openai: request failed: %w", err)
	}
	defer func() {
		// Drain body to enable connection reuse
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
		resp.Body.Close()
	}()

	// Read response body with size limit
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("ai/openai: failed to read response: %w", err)
	}

	// Handle HTTP errors
	if err := p.handleHTTPError(resp.StatusCode, respBody); err != nil {
		return ChatResponse{}, err
	}

	// Parse response
	return p.parseResponse(respBody)
}

// buildRequestPayload constructs the OpenAI API request payload.
func (p *OpenAIProvider) buildRequestPayload(req ChatRequest) openAIChatRequest {
	// Determine effective parameters
	systemMessage := req.System
	if systemMessage == "" {
		systemMessage = p.defaultSystem
	}

	temperature := p.defaultTemperature
	if req.Temperature != nil {
		temperature = *req.Temperature
	}

	maxTokens := p.defaultMaxTokens
	if req.MaxTokens != nil {
		maxTokens = *req.MaxTokens
	}

	payload := openAIChatRequest{
		Model: p.model,
		Messages: []openAIChatMessage{
			{Role: "system", Content: systemMessage},
			{Role: "user", Content: req.Prompt},
		},
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}

	if req.TopP != nil {
		payload.TopP = req.TopP
	}

	if len(req.Stop) > 0 {
		payload.Stop = req.Stop
	}

	if req.User != "" {
		payload.User = req.User
	}

	return payload
}

// handleHTTPError translates HTTP status codes to appropriate errors.
func (p *OpenAIProvider) handleHTTPError(statusCode int, body []byte) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}

	// Try to parse error response
	var apiResp openAIChatResponse
	if err := json.Unmarshal(body, &apiResp); err == nil && apiResp.Error != nil {
		errMsg := fmt.Sprintf("ai/openai: API error (status %d): %s", statusCode, apiResp.Error.Message)
		if apiResp.Error.Code != "" {
			errMsg += fmt.Sprintf(" [code: %s]", apiResp.Error.Code)
		}

		// Map to sentinel errors where appropriate
		switch statusCode {
		case http.StatusTooManyRequests:
			return fmt.Errorf("%w: %s", ErrOpenAIRateLimited, apiResp.Error.Message)
		case http.StatusUnauthorized:
			return fmt.Errorf("%w: invalid API key", ErrOpenAINotConfigured)
		}

		if statusCode >= 500 {
			return fmt.Errorf("%w: %s", ErrOpenAIServerError, apiResp.Error.Message)
		}

		return errors.New(errMsg)
	}

	// Generic error if we can't parse the response
	switch statusCode {
	case http.StatusTooManyRequests:
		return ErrOpenAIRateLimited
	case http.StatusUnauthorized:
		return fmt.Errorf("%w: invalid API key", ErrOpenAINotConfigured)
	}

	if statusCode >= 500 {
		return fmt.Errorf("%w: status %d", ErrOpenAIServerError, statusCode)
	}

	return fmt.Errorf("ai/openai: unexpected status %d", statusCode)
}

// parseResponse parses the OpenAI API response into a ChatResponse.
func (p *OpenAIProvider) parseResponse(body []byte) (ChatResponse, error) {
	var apiResp openAIChatResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return ChatResponse{}, fmt.Errorf("%w: %v", ErrOpenAIInvalidResponse, err)
	}

	if len(apiResp.Choices) == 0 {
		return ChatResponse{}, ErrOpenAINoChoices
	}

	choice := apiResp.Choices[0]

	// Check for content filtering
	if choice.FinishReason == "content_filter" {
		return ChatResponse{}, ErrOpenAIContentFiltered
	}

	return ChatResponse{
		Output:       choice.Message.Content,
		Source:       ChatSourceCloud,
		Model:        apiResp.Model,
		FinishReason: choice.FinishReason,
		RequestID:    apiResp.ID,
		CreatedAt:    time.Unix(apiResp.Created, 0),
		Usage: &TokenUsage{
			PromptTokens:     apiResp.Usage.PromptTokens,
			CompletionTokens: apiResp.Usage.CompletionTokens,
			TotalTokens:      apiResp.Usage.TotalTokens,
		},
		ProviderMetadata: map[string]any{
			"openai_id":          apiResp.ID,
			"openai_object":      apiResp.Object,
			"system_fingerprint": apiResp.SystemFingerprint,
		},
	}, nil
}

// IsConfigured returns true if the provider has a valid API key.
func (p *OpenAIProvider) IsConfigured() bool {
	return p != nil && strings.TrimSpace(p.apiKey) != ""
}

// HealthCheck performs a lightweight check of OpenAI API availability.
// It uses the models endpoint which is fast and doesn't consume tokens.
func (p *OpenAIProvider) HealthCheck(ctx context.Context) error {
	if !p.IsConfigured() {
		return ErrOpenAINotConfigured
	}

	endpoint := p.baseURL + "/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("ai/openai: health check failed: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ai/openai: health check failed: %w", err)
	}
	defer resp.Body.Close()

	// Drain body
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ai/openai: health check returned status %d", resp.StatusCode)
	}

	return nil
}

// Metrics returns current provider metrics.
func (p *OpenAIProvider) Metrics() ProviderMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var avgLatency time.Duration
	if p.requestCount > 0 {
		avgLatency = p.totalLatency / time.Duration(p.requestCount)
	}

	return ProviderMetrics{
		RequestCount:   p.requestCount,
		ErrorCount:     p.errorCount,
		AverageLatency: avgLatency,
	}
}

// ProviderMetrics contains runtime metrics for the provider.
type ProviderMetrics struct {
	RequestCount   int64
	ErrorCount     int64
	AverageLatency time.Duration
}

// =============================================================================
// OpenAI API Types
// =============================================================================

// openAIChatRequest is the request payload for OpenAI chat completions.
type openAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []openAIChatMessage `json:"messages"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
	Temperature float64             `json:"temperature,omitempty"`
	TopP        *float64            `json:"top_p,omitempty"`
	Stop        []string            `json:"stop,omitempty"`
	User        string              `json:"user,omitempty"`
}

// openAIChatMessage represents a single message in the conversation.
type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openAIChatResponse is the response from OpenAI chat completions.
type openAIChatResponse struct {
	ID                string             `json:"id"`
	Object            string             `json:"object"`
	Created           int64              `json:"created"`
	Model             string             `json:"model"`
	SystemFingerprint string             `json:"system_fingerprint,omitempty"`
	Choices           []openAIChatChoice `json:"choices"`
	Usage             openAIUsage        `json:"usage"`
	Error             *openAIError       `json:"error,omitempty"`
}

// openAIChatChoice represents a single completion choice.
type openAIChatChoice struct {
	Index        int               `json:"index"`
	Message      openAIChatMessage `json:"message"`
	FinishReason string            `json:"finish_reason"`
}

// openAIUsage describes token usage information.
type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// openAIError represents an error from the OpenAI API.
type openAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
	Param   string `json:"param,omitempty"`
}

// =============================================================================
// Compile-time interface checks
// =============================================================================

var (
	_ CloudProvider = (*OpenAIProvider)(nil)
	_ HealthChecker = (*OpenAIProvider)(nil)
)
