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
	"sync"
	"time"
)

// LocalOfflineProvider implements a local AI inference engine using llama.cpp or Ollama
type LocalOfflineProvider struct {
	baseURL    string
	model      string
	httpClient *http.Client
	available  bool
	mu         sync.RWMutex
}

// LocalOfflineConfig holds configuration for local offline AI
type LocalOfflineConfig struct {
	BaseURL string
	Model   string
	Timeout time.Duration
}

// NewLocalOfflineProvider creates a local offline AI provider
func NewLocalOfflineProvider(config LocalOfflineConfig) (*LocalOfflineProvider, error) {
	if config.BaseURL == "" {
		// Try to detect from environment
		if ollamaURL := os.Getenv("OLLAMA_HOST"); ollamaURL != "" {
			config.BaseURL = ollamaURL
		} else {
			config.BaseURL = "http://localhost:11434"
		}
	}

	if config.Model == "" {
		config.Model = "llama3.2:3b" // Lightweight model for edge devices
	}

	if config.Timeout == 0 {
		config.Timeout = 120 * time.Second // Local inference can take longer
	}

	provider := &LocalOfflineProvider{
		baseURL: config.BaseURL,
		model:   config.Model,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}

	// Check availability on initialization
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	provider.available = provider.checkAvailability(ctx)

	return provider, nil
}

// NewLocalOfflineProviderFromEnv creates a provider from environment variables
func NewLocalOfflineProviderFromEnv() (*LocalOfflineProvider, error) {
	return NewLocalOfflineProvider(LocalOfflineConfig{
		BaseURL: os.Getenv("OFFGRIDFLOW_LOCAL_AI_URL"),
		Model:   os.Getenv("OFFGRIDFLOW_LOCAL_AI_MODEL"),
	})
}

// Chat sends a request to the local AI engine
func (p *LocalOfflineProvider) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	startTime := time.Now()

	if err := req.Validate(); err != nil {
		return ChatResponse{}, err
	}

	if !p.IsAvailable() {
		return ChatResponse{}, ErrProviderUnavailable
	}

	// Build request for Ollama/llama.cpp
	reqPayload := map[string]interface{}{
		"model":  p.model,
		"prompt": p.buildPrompt(req),
		"stream": false,
	}

	if req.Temperature != nil {
		reqPayload["temperature"] = *req.Temperature
	}

	if req.MaxTokens != nil {
		reqPayload["max_tokens"] = *req.MaxTokens
	}

	body, err := json.Marshal(reqPayload)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("local-ai: failed to marshal request: %w", err)
	}

	// Send request
	endpoint := p.baseURL + "/api/generate"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("local-ai: failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		p.markUnavailable()
		if ctx.Err() != nil {
			return ChatResponse{}, fmt.Errorf("local-ai: %w", ctx.Err())
		}
		return ChatResponse{}, fmt.Errorf("local-ai: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return ChatResponse{}, fmt.Errorf("local-ai: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
		Context  []int  `json:"context,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ChatResponse{}, fmt.Errorf("local-ai: failed to decode response: %w", err)
	}

	if !result.Done {
		return ChatResponse{}, errors.New("local-ai: incomplete response")
	}

	return ChatResponse{
		Output:       result.Response,
		Source:       ChatSourceLocal,
		Model:        p.model,
		FinishReason: "stop",
		Latency:      time.Since(startTime),
		CreatedAt:    time.Now(),
	}, nil
}

// buildPrompt constructs a prompt for the local model
func (p *LocalOfflineProvider) buildPrompt(req ChatRequest) string {
	systemPrompt := req.System
	if systemPrompt == "" {
		systemPrompt = defaultLocalSystemPrompt
	}

	return fmt.Sprintf("System: %s\n\nUser: %s\n\nAssistant:", systemPrompt, req.Prompt)
}

const defaultLocalSystemPrompt = `You are a helpful AI assistant for OffGridFlow, a carbon accounting platform.
Provide concise, accurate responses about emissions calculations, sustainability, and carbon reporting.
Focus on practical advice and calculations.`

// IsAvailable checks if the local AI engine is available
func (p *LocalOfflineProvider) IsAvailable() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.available
}

// checkAvailability pings the local AI endpoint
func (p *LocalOfflineProvider) checkAvailability(ctx context.Context) bool {
	endpoint := p.baseURL + "/api/tags"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return false
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// markUnavailable marks the provider as unavailable
func (p *LocalOfflineProvider) markUnavailable() {
	p.mu.Lock()
	p.available = false
	p.mu.Unlock()
}

// HealthCheck performs a health check on the local AI engine
func (p *LocalOfflineProvider) HealthCheck(ctx context.Context) error {
	if p.checkAvailability(ctx) {
		p.mu.Lock()
		p.available = true
		p.mu.Unlock()
		return nil
	}

	p.markUnavailable()
	return fmt.Errorf("local-ai: engine not available at %s", p.baseURL)
}

// PullModel downloads a model to the local AI engine
func (p *LocalOfflineProvider) PullModel(ctx context.Context, modelName string) error {
	endpoint := p.baseURL + "/api/pull"

	reqPayload := map[string]string{
		"name": modelName,
	}

	body, err := json.Marshal(reqPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal pull request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to pull model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("pull failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListModels returns available models on the local AI engine
func (p *LocalOfflineProvider) ListModels(ctx context.Context) ([]string, error) {
	endpoint := p.baseURL + "/api/tags"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list models failed with status %d", resp.StatusCode)
	}

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]string, len(result.Models))
	for i, m := range result.Models {
		models[i] = m.Name
	}

	return models, nil
}

// Close releases resources
func (p *LocalOfflineProvider) Close() error {
	p.httpClient.CloseIdleConnections()
	return nil
}

// Compile-time interface check
var (
	_ LocalProvider = (*LocalOfflineProvider)(nil)
	_ HealthChecker = (*LocalOfflineProvider)(nil)
	_ Closer        = (*LocalOfflineProvider)(nil)
)
