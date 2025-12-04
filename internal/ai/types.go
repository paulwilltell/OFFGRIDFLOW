// Package ai provides a unified interface for AI chat providers with support
// for both cloud-based and local inference engines. It implements an intelligent
// routing system that seamlessly handles online/offline mode transitions and
// automatic failover between providers.
//
// Architecture Overview:
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│                         Router                                  │
//	│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐        │
//	│  │ ModeManager │───▶│   Route     │───▶│  Response   │        │
//	│  └─────────────┘    │  Decision   │    └─────────────┘        │
//	│                      └──────┬──────┘                           │
//	│                             │                                   │
//	│              ┌──────────────┴──────────────┐                   │
//	│              ▼                              ▼                   │
//	│     ┌─────────────┐                ┌─────────────┐            │
//	│     │    Cloud    │                │    Local    │            │
//	│     │  Provider   │                │  Provider   │            │
//	│     └─────────────┘                └─────────────┘            │
//	└─────────────────────────────────────────────────────────────────┘
//
// Usage:
//
//	router := ai.NewRouter(modeManager, cloudProvider, localProvider)
//	resp, err := router.Chat(ctx, ai.ChatRequest{Prompt: "Hello"})
package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// =============================================================================
// Sentinel Errors
// =============================================================================

var (
	// ErrEmptyPrompt is returned when a ChatRequest contains an empty or
	// whitespace-only prompt. All providers must validate this condition
	// before processing the request.
	ErrEmptyPrompt = errors.New("ai: prompt cannot be empty")

	// ErrProviderUnavailable indicates the target provider cannot service
	// requests at this time. This may be due to configuration issues,
	// network problems, or rate limiting.
	ErrProviderUnavailable = errors.New("ai: provider unavailable")

	// ErrInvalidConfiguration indicates the provider was not properly
	// configured with required parameters (e.g., missing API key).
	ErrInvalidConfiguration = errors.New("ai: invalid provider configuration")

	// ErrRateLimited indicates the provider has rate-limited the request.
	// Callers should implement exponential backoff before retrying.
	ErrRateLimited = errors.New("ai: rate limited by provider")

	// ErrContextCanceled wraps context cancellation for consistent error handling.
	ErrContextCanceled = errors.New("ai: context canceled")

	// ErrMaxTokensExceeded indicates the request would exceed the maximum
	// token limit for the model.
	ErrMaxTokensExceeded = errors.New("ai: maximum tokens exceeded")

	// ErrContentFiltered indicates the request or response was blocked by
	// content safety filters.
	ErrContentFiltered = errors.New("ai: content filtered by safety policy")

	// ErrModelNotFound indicates the requested model does not exist or is
	// not available in the current configuration.
	ErrModelNotFound = errors.New("ai: model not found")

	// ErrTimeout indicates the request exceeded the configured timeout.
	ErrTimeout = errors.New("ai: request timeout")
)

// =============================================================================
// Chat Source Enumeration
// =============================================================================

// ChatSource identifies the origin of a ChatResponse, enabling callers to
// understand which provider handled their request and make decisions based
// on the response source (e.g., caching strategies, telemetry).
type ChatSource string

const (
	// ChatSourceCloud indicates the response originated from a cloud-based
	// AI provider (e.g., OpenAI, Anthropic, Azure OpenAI).
	ChatSourceCloud ChatSource = "cloud"

	// ChatSourceLocal indicates the response originated from a locally-running
	// model or inference engine (e.g., Ollama, llama.cpp, ONNX runtime).
	ChatSourceLocal ChatSource = "local"

	// ChatSourceCache indicates the response was served from a cache layer.
	// This source is used when semantic caching is enabled.
	ChatSourceCache ChatSource = "cache"

	// ChatSourceFallback indicates the response came from a fallback mechanism
	// after the primary provider failed.
	ChatSourceFallback ChatSource = "fallback"
)

// String returns a human-readable representation of the chat source.
func (cs ChatSource) String() string {
	return string(cs)
}

// IsValid returns true if the ChatSource is a recognized value.
func (cs ChatSource) IsValid() bool {
	switch cs {
	case ChatSourceCloud, ChatSourceLocal, ChatSourceCache, ChatSourceFallback:
		return true
	default:
		return false
	}
}

// MarshalText implements encoding.TextMarshaler for ChatSource.
func (cs ChatSource) MarshalText() ([]byte, error) {
	return []byte(cs), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for ChatSource.
func (cs *ChatSource) UnmarshalText(text []byte) error {
	source := ChatSource(text)
	if !source.IsValid() {
		return fmt.Errorf("ai: invalid chat source: %q", text)
	}
	*cs = source
	return nil
}

// =============================================================================
// Request Types
// =============================================================================

// ChatRequest encapsulates all parameters for an AI chat completion request.
// It provides a provider-agnostic interface that can be translated to any
// underlying API format (OpenAI, Anthropic, local models, etc.).
//
// Required fields:
//   - Prompt: The user's message (must be non-empty)
//
// Optional fields:
//   - System: Custom system instructions (provider default used if empty)
//   - Temperature: Sampling temperature (provider default used if nil)
//   - MaxTokens: Maximum response tokens (provider default used if nil)
//   - TopP: Nucleus sampling parameter (provider default used if nil)
//   - Metadata: Arbitrary key-value pairs for tracking/logging
type ChatRequest struct {
	// Prompt is the primary user message that the AI will respond to.
	// This field is required and must contain non-whitespace content.
	Prompt string `json:"prompt" validate:"required,min=1"`

	// System provides optional system-level instructions that guide the
	// assistant's behavior and persona. When empty, the provider's
	// default system prompt is used.
	//
	// Example: "You are a helpful carbon accounting expert."
	System string `json:"system,omitempty"`

	// Temperature controls the randomness of the output. Values range
	// from 0.0 (deterministic) to 2.0 (highly creative). When nil,
	// the provider's default temperature is used.
	//
	// Recommended values:
	//   - 0.0-0.3: Factual, consistent responses
	//   - 0.5-0.7: Balanced creativity and coherence
	//   - 0.8-1.0: Creative, varied responses
	Temperature *float64 `json:"temperature,omitempty" validate:"omitempty,gte=0,lte=2"`

	// MaxTokens sets an upper bound on the number of tokens in the
	// generated response. When nil, the provider's default limit is used.
	// Note: 1 token ≈ 4 characters in English text.
	MaxTokens *int `json:"max_tokens,omitempty" validate:"omitempty,gte=1"`

	// TopP implements nucleus sampling. The model considers tokens
	// comprising the top P probability mass. When nil, the provider's
	// default is used. Typically, only one of Temperature or TopP
	// should be modified, not both.
	TopP *float64 `json:"top_p,omitempty" validate:"omitempty,gte=0,lte=1"`

	// Stop specifies up to 4 sequences where the API will stop generating
	// further tokens. The returned text will not contain the stop sequence.
	Stop []string `json:"stop,omitempty" validate:"omitempty,max=4"`

	// Metadata contains arbitrary key-value pairs for request tracking,
	// logging, or passing through to provider-specific features.
	// This field is not sent to providers but is available in responses.
	Metadata map[string]any `json:"metadata,omitempty"`

	// ConversationID links this request to a conversation for multi-turn
	// context. Providers that support conversation state will use this
	// to retrieve prior messages.
	ConversationID string `json:"conversation_id,omitempty"`

	// User is an optional identifier for the end-user making the request.
	// This can be used for abuse detection and monitoring.
	User string `json:"user,omitempty"`
}

// Validate performs validation on the ChatRequest and returns an error
// if any required fields are missing or invalid.
func (r ChatRequest) Validate() error {
	if strings.TrimSpace(r.Prompt) == "" {
		return ErrEmptyPrompt
	}

	if r.Temperature != nil {
		if *r.Temperature < 0 || *r.Temperature > 2 {
			return fmt.Errorf("ai: temperature must be between 0 and 2, got %f", *r.Temperature)
		}
	}

	if r.MaxTokens != nil && *r.MaxTokens < 1 {
		return fmt.Errorf("ai: max_tokens must be positive, got %d", *r.MaxTokens)
	}

	if r.TopP != nil {
		if *r.TopP < 0 || *r.TopP > 1 {
			return fmt.Errorf("ai: top_p must be between 0 and 1, got %f", *r.TopP)
		}
	}

	if len(r.Stop) > 4 {
		return fmt.Errorf("ai: stop sequences limited to 4, got %d", len(r.Stop))
	}

	return nil
}

// WithTemperature returns a copy of the request with the specified temperature.
// This enables fluent request building.
func (r ChatRequest) WithTemperature(temp float64) ChatRequest {
	r.Temperature = &temp
	return r
}

// WithMaxTokens returns a copy of the request with the specified max tokens.
// This enables fluent request building.
func (r ChatRequest) WithMaxTokens(tokens int) ChatRequest {
	r.MaxTokens = &tokens
	return r
}

// WithSystem returns a copy of the request with the specified system prompt.
// This enables fluent request building.
func (r ChatRequest) WithSystem(system string) ChatRequest {
	r.System = system
	return r
}

// WithTopP returns a copy of the request with the specified top_p value.
// This enables fluent request building.
func (r ChatRequest) WithTopP(topP float64) ChatRequest {
	r.TopP = &topP
	return r
}

// WithMetadata returns a copy of the request with the specified metadata.
// This enables fluent request building.
func (r ChatRequest) WithMetadata(key string, value any) ChatRequest {
	if r.Metadata == nil {
		r.Metadata = make(map[string]any)
	}
	r.Metadata[key] = value
	return r
}

// Clone creates a deep copy of the ChatRequest.
func (r ChatRequest) Clone() ChatRequest {
	clone := r

	// Deep copy pointer fields
	if r.Temperature != nil {
		temp := *r.Temperature
		clone.Temperature = &temp
	}
	if r.MaxTokens != nil {
		tokens := *r.MaxTokens
		clone.MaxTokens = &tokens
	}
	if r.TopP != nil {
		topP := *r.TopP
		clone.TopP = &topP
	}

	// Deep copy slices
	if r.Stop != nil {
		clone.Stop = make([]string, len(r.Stop))
		copy(clone.Stop, r.Stop)
	}

	// Deep copy maps
	if r.Metadata != nil {
		clone.Metadata = make(map[string]any, len(r.Metadata))
		for k, v := range r.Metadata {
			clone.Metadata[k] = v
		}
	}

	return clone
}

// =============================================================================
// Response Types
// =============================================================================

// ChatResponse encapsulates the result of an AI chat completion request.
// It provides a normalized interface regardless of which provider generated
// the response.
type ChatResponse struct {
	// Output is the text content of the assistant's reply.
	// This is the primary payload of the response.
	Output string `json:"output"`

	// Source identifies which provider path produced this response,
	// enabling callers to implement source-specific logic.
	Source ChatSource `json:"source"`

	// Model identifies the specific model that generated this response
	// (e.g., "gpt-4o-mini", "claude-3-sonnet").
	Model string `json:"model,omitempty"`

	// Usage contains token consumption metrics for the request.
	// This may be nil if the provider doesn't report usage.
	Usage *TokenUsage `json:"usage,omitempty"`

	// FinishReason indicates why the model stopped generating tokens.
	// Common values: "stop", "length", "content_filter", "tool_calls"
	FinishReason string `json:"finish_reason,omitempty"`

	// ProviderMetadata contains provider-specific details that don't fit
	// the normalized schema. This field is intentionally excluded from
	// JSON serialization to avoid leaking implementation details.
	ProviderMetadata any `json:"-"`

	// Latency records the time taken to generate this response.
	// This is measured from request initiation to response completion.
	Latency time.Duration `json:"latency_ns,omitempty"`

	// RequestID is a unique identifier for this request, useful for
	// debugging and correlating with provider logs.
	RequestID string `json:"request_id,omitempty"`

	// CreatedAt records when the response was generated.
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// TokenUsage provides detailed token consumption metrics for a request.
// This information is valuable for cost tracking and quota management.
type TokenUsage struct {
	// PromptTokens is the number of tokens in the input prompt.
	PromptTokens int `json:"prompt_tokens"`

	// CompletionTokens is the number of tokens in the generated response.
	CompletionTokens int `json:"completion_tokens"`

	// TotalTokens is the sum of PromptTokens and CompletionTokens.
	TotalTokens int `json:"total_tokens"`

	// CachedTokens indicates tokens served from provider-side caching.
	// This reduces cost for repeated context in multi-turn conversations.
	CachedTokens int `json:"cached_tokens,omitempty"`
}

// IsEmpty returns true if the response contains no meaningful output.
func (r ChatResponse) IsEmpty() bool {
	return strings.TrimSpace(r.Output) == ""
}

// TruncatedOutput returns the output truncated to the specified length
// with an ellipsis appended if truncation occurred.
func (r ChatResponse) TruncatedOutput(maxLen int) string {
	if len(r.Output) <= maxLen {
		return r.Output
	}
	if maxLen <= 3 {
		return r.Output[:maxLen]
	}
	return r.Output[:maxLen-3] + "..."
}

// =============================================================================
// Provider Interfaces
// =============================================================================

// CloudProvider defines the interface for cloud-based AI services such as
// OpenAI, Anthropic, Azure OpenAI, or Google Vertex AI. Implementations
// must be safe for concurrent use.
//
// Cloud providers typically offer:
//   - Latest model capabilities
//   - High availability and scalability
//   - Usage-based pricing
//   - Managed infrastructure
//
// Implementations should handle:
//   - Authentication and API key management
//   - Rate limiting and retry logic
//   - Request/response serialization
//   - Error translation to package-level errors
type CloudProvider interface {
	// Chat sends a request to the cloud AI service and returns the response.
	// The context should be used for cancellation and timeout control.
	//
	// Errors:
	//   - ErrEmptyPrompt: prompt was empty or whitespace-only
	//   - ErrProviderUnavailable: service is unreachable
	//   - ErrRateLimited: request was rate-limited
	//   - ErrInvalidConfiguration: provider not properly configured
	//   - context.Canceled/DeadlineExceeded: request was canceled
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)

	// IsConfigured returns true if the provider has valid configuration
	// (e.g., API key is set). This enables graceful degradation when
	// cloud providers are optional.
	IsConfigured() bool
}

// LocalProvider defines the interface for locally-running AI inference.
// This includes embedded models, local LLM servers (Ollama, llama.cpp),
// or rule-based fallback systems.
//
// Local providers are essential for:
//   - Offline operation (air-gapped environments)
//   - Data privacy (no external API calls)
//   - Cost reduction (no per-token charges)
//   - Low latency (no network round-trip)
//
// Implementations should handle:
//   - Model loading and initialization
//   - Resource management (memory, GPU)
//   - Graceful degradation when resources are constrained
type LocalProvider interface {
	// Chat generates a response using local inference.
	// The context should be used for cancellation and timeout control.
	//
	// Errors:
	//   - ErrEmptyPrompt: prompt was empty or whitespace-only
	//   - ErrProviderUnavailable: local model not loaded/available
	//   - context.Canceled/DeadlineExceeded: request was canceled
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)

	// IsAvailable returns true if the local provider is ready to serve
	// requests (model loaded, resources available).
	IsAvailable() bool
}

// =============================================================================
// Optional Provider Interfaces
// =============================================================================

// HealthChecker is an optional interface that providers can implement
// to expose health status for monitoring and load balancing.
type HealthChecker interface {
	// HealthCheck performs a lightweight check of provider availability.
	// It should complete quickly and not consume significant resources.
	HealthCheck(ctx context.Context) error
}

// Closer is an optional interface for providers that hold resources
// and need cleanup.
type Closer interface {
	// Close releases any resources held by the provider.
	Close() error
}

// StreamingProvider is an optional interface for providers that support
// streaming responses token-by-token.
type StreamingProvider interface {
	// ChatStream sends a request and returns a channel of response chunks.
	// The channel is closed when the response is complete or an error occurs.
	ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
}

// StreamChunk represents a single chunk in a streaming response.
type StreamChunk struct {
	// Content is the text content of this chunk.
	Content string `json:"content"`

	// Done indicates this is the final chunk.
	Done bool `json:"done"`

	// Error contains any error that occurred during streaming.
	Error error `json:"-"`
}

// =============================================================================
// Helper Functions
// =============================================================================

// NewChatRequest creates a new ChatRequest with the given prompt.
// This is a convenience function for simple single-turn requests.
func NewChatRequest(prompt string) ChatRequest {
	return ChatRequest{Prompt: prompt}
}

// Ptr returns a pointer to the given value. This is a convenience function
// for setting optional pointer fields in ChatRequest.
func Ptr[T any](v T) *T {
	return &v
}

// WrapError wraps an error with additional context while preserving
// the ability to use errors.Is and errors.As.
func WrapError(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// IsRetryableError returns true if the error is transient and the request
// should be retried after a backoff period.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for known retryable errors
	if errors.Is(err, ErrRateLimited) ||
		errors.Is(err, ErrTimeout) ||
		errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Check for network-related errors (these often contain "timeout" or "connection")
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "temporary failure")
}
