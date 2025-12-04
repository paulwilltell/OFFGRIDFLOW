package ai

import (
	"context"
	"strings"
	"time"
)

// =============================================================================
// Stub Cloud Provider
// =============================================================================

// StubCloudProvider is a minimal CloudProvider implementation intended for
// unit tests, integration tests, and local development environments where
// real API calls are not desired or possible.
//
// This stub provides deterministic, predictable responses that make tests
// reproducible and fast. It never makes external network calls.
//
// Usage:
//
//	stub := &ai.StubCloudProvider{}
//	resp, err := stub.Chat(ctx, ai.ChatRequest{Prompt: "Hello"})
//	// resp.Output == "[CLOUD STUB] Hello"
//	// resp.Source == ai.ChatSourceCloud
type StubCloudProvider struct {
	// ResponsePrefix is prepended to the prompt in responses.
	// Defaults to "[CLOUD STUB] " if empty.
	ResponsePrefix string

	// SimulatedLatency adds artificial delay to simulate network latency.
	// Useful for testing timeout handling and loading states.
	SimulatedLatency time.Duration

	// FailWithError, when non-nil, causes Chat to return this error.
	// Useful for testing error handling and fallback behavior.
	FailWithError error

	// CallCount tracks the number of times Chat has been called.
	// Useful for verifying expected call patterns in tests.
	CallCount int

	// LastRequest stores the most recent request for inspection.
	LastRequest *ChatRequest

	// CustomHandler, when set, is called instead of the default logic.
	// This enables complex test scenarios with conditional responses.
	CustomHandler func(ctx context.Context, req ChatRequest) (ChatResponse, error)
}

// Chat returns a deterministic echo-style response that identifies itself
// as originating from the cloud stub. This implementation is thread-safe
// for reads but not for concurrent writes to CallCount/LastRequest.
func (s *StubCloudProvider) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	s.CallCount++
	reqCopy := req.Clone()
	s.LastRequest = &reqCopy

	// Check for context cancellation first
	if err := ctx.Err(); err != nil {
		return ChatResponse{}, err
	}

	// Use custom handler if provided
	if s.CustomHandler != nil {
		return s.CustomHandler(ctx, req)
	}

	// Return configured error if set
	if s.FailWithError != nil {
		return ChatResponse{}, s.FailWithError
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return ChatResponse{}, err
	}

	// Simulate network latency if configured
	if s.SimulatedLatency > 0 {
		select {
		case <-time.After(s.SimulatedLatency):
		case <-ctx.Done():
			return ChatResponse{}, ctx.Err()
		}
	}

	prefix := s.ResponsePrefix
	if prefix == "" {
		prefix = "[CLOUD STUB] "
	}

	return ChatResponse{
		Output:       prefix + req.Prompt,
		Source:       ChatSourceCloud,
		Model:        "stub-cloud-v1",
		FinishReason: "stop",
		CreatedAt:    time.Now(),
		Usage: &TokenUsage{
			PromptTokens:     len(strings.Fields(req.Prompt)),
			CompletionTokens: len(strings.Fields(req.Prompt)) + 2,
			TotalTokens:      len(strings.Fields(req.Prompt))*2 + 2,
		},
	}, nil
}

// IsConfigured always returns true for the stub provider.
func (s *StubCloudProvider) IsConfigured() bool {
	return true
}

// Reset clears all recorded state (call count, last request).
// Useful for reusing the same stub across multiple test cases.
func (s *StubCloudProvider) Reset() {
	s.CallCount = 0
	s.LastRequest = nil
}

// =============================================================================
// Stub Local Provider
// =============================================================================

// StubLocalProvider is a minimal LocalProvider implementation intended for
// unit tests, integration tests, and local development environments.
//
// This stub mirrors StubCloudProvider but identifies responses as coming
// from the local inference path.
//
// Usage:
//
//	stub := &ai.StubLocalProvider{}
//	resp, err := stub.Chat(ctx, ai.ChatRequest{Prompt: "Hello"})
//	// resp.Output == "[LOCAL STUB] Hello"
//	// resp.Source == ai.ChatSourceLocal
type StubLocalProvider struct {
	// ResponsePrefix is prepended to the prompt in responses.
	// Defaults to "[LOCAL STUB] " if empty.
	ResponsePrefix string

	// SimulatedLatency adds artificial delay to simulate inference time.
	SimulatedLatency time.Duration

	// FailWithError, when non-nil, causes Chat to return this error.
	FailWithError error

	// Available controls the return value of IsAvailable().
	// Defaults to true if not explicitly set.
	Available *bool

	// CallCount tracks the number of times Chat has been called.
	CallCount int

	// LastRequest stores the most recent request for inspection.
	LastRequest *ChatRequest

	// CustomHandler, when set, is called instead of the default logic.
	CustomHandler func(ctx context.Context, req ChatRequest) (ChatResponse, error)
}

// Chat returns a deterministic echo-style response that identifies itself
// as originating from the local stub.
func (s *StubLocalProvider) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	s.CallCount++
	reqCopy := req.Clone()
	s.LastRequest = &reqCopy

	// Check for context cancellation first
	if err := ctx.Err(); err != nil {
		return ChatResponse{}, err
	}

	// Use custom handler if provided
	if s.CustomHandler != nil {
		return s.CustomHandler(ctx, req)
	}

	// Return configured error if set
	if s.FailWithError != nil {
		return ChatResponse{}, s.FailWithError
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return ChatResponse{}, err
	}

	// Check availability
	if !s.IsAvailable() {
		return ChatResponse{}, ErrProviderUnavailable
	}

	// Simulate inference time if configured
	if s.SimulatedLatency > 0 {
		select {
		case <-time.After(s.SimulatedLatency):
		case <-ctx.Done():
			return ChatResponse{}, ctx.Err()
		}
	}

	prefix := s.ResponsePrefix
	if prefix == "" {
		prefix = "[LOCAL STUB] "
	}

	return ChatResponse{
		Output:       prefix + req.Prompt,
		Source:       ChatSourceLocal,
		Model:        "stub-local-v1",
		FinishReason: "stop",
		CreatedAt:    time.Now(),
	}, nil
}

// IsAvailable returns the configured availability state, defaulting to true.
func (s *StubLocalProvider) IsAvailable() bool {
	if s.Available == nil {
		return true
	}
	return *s.Available
}

// Reset clears all recorded state.
func (s *StubLocalProvider) Reset() {
	s.CallCount = 0
	s.LastRequest = nil
}

// SetAvailable configures the availability state.
func (s *StubLocalProvider) SetAvailable(available bool) {
	s.Available = &available
}

// =============================================================================
// Compile-time interface checks
// =============================================================================

var (
	_ CloudProvider = (*StubCloudProvider)(nil)
	_ LocalProvider = (*StubLocalProvider)(nil)
)
