package models

// ChatMessage represents a single message in a conversation.
// Content is string-only in this mock (multimodal content parts are out of scope).
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// ChatStreamOptions controls the trailing usage chunk when streaming.
type ChatStreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

// ChatRequest is the body sent by the client to POST /v1/chat/completions.
type ChatRequest struct {
	Model               string             `json:"model"`
	Messages            []ChatMessage      `json:"messages"`
	Stream              bool               `json:"stream,omitempty"`
	StreamOptions       *ChatStreamOptions `json:"stream_options,omitempty"`
	Temperature         *float64           `json:"temperature,omitempty"`
	TopP                *float64           `json:"top_p,omitempty"`
	N                   *int               `json:"n,omitempty"`
	Stop                interface{}        `json:"stop,omitempty"`
	MaxTokens           *int               `json:"max_tokens,omitempty"` // deprecated, kept for compatibility
	MaxCompletionTokens *int               `json:"max_completion_tokens,omitempty"`
	PresencePenalty     *float64           `json:"presence_penalty,omitempty"`
	FrequencyPenalty    *float64           `json:"frequency_penalty,omitempty"`
	Logprobs            bool               `json:"logprobs,omitempty"`
	TopLogprobs         *int               `json:"top_logprobs,omitempty"`
	Seed                *int64             `json:"seed,omitempty"`
	User                string             `json:"user,omitempty"`
}

// ChatResponseMessage is the assistant message returned in a completion choice.
type ChatResponseMessage struct {
	Role        string        `json:"role"`
	Content     *string       `json:"content"`
	Refusal     *string       `json:"refusal"`
	Annotations []interface{} `json:"annotations"`
}

// ChatChoice represents one possible response choice.
type ChatChoice struct {
	Index        int                 `json:"index"`
	Message      ChatResponseMessage `json:"message"`
	Logprobs     interface{}         `json:"logprobs"`
	FinishReason string              `json:"finish_reason"`
}

// TokenDetails breakdowns matching CompletionUsage in openai.yml.
type PromptTokensDetails struct {
	CachedTokens int `json:"cached_tokens"`
	AudioTokens  int `json:"audio_tokens"`
}

type CompletionTokensDetails struct {
	ReasoningTokens          int `json:"reasoning_tokens"`
	AudioTokens              int `json:"audio_tokens"`
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
	RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
}

// UsageStats tracks token consumption (CompletionUsage schema).
type UsageStats struct {
	PromptTokens            int                     `json:"prompt_tokens"`
	CompletionTokens        int                     `json:"completion_tokens"`
	TotalTokens             int                     `json:"total_tokens"`
	PromptTokensDetails     PromptTokensDetails     `json:"prompt_tokens_details"`
	CompletionTokensDetails CompletionTokensDetails `json:"completion_tokens_details"`
}

// ChatResponse is the full JSON object returned to the client.
type ChatResponse struct {
	ID                string            `json:"id"`
	Object            string            `json:"object"`
	Created           int64             `json:"created"`
	Model             string            `json:"model"`
	Choices           []ChatChoice      `json:"choices"`
	Usage             UsageStats        `json:"usage"`
	ServiceTier       string            `json:"service_tier,omitempty"`
	SystemFingerprint string            `json:"system_fingerprint,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty"`

	// Not part of the OpenAI response shape; kept only inside the
	// in-memory store to answer GET .../messages for a stored completion.
	RequestMessages []ChatMessage `json:"-"`
}

// --- Streaming (chat.completion.chunk) ---

type ChatStreamDelta struct {
	Role    string  `json:"role,omitempty"`
	Content *string `json:"content,omitempty"`
}

type ChatStreamChoice struct {
	Index        int             `json:"index"`
	Delta        ChatStreamDelta `json:"delta"`
	Logprobs     interface{}     `json:"logprobs"`
	FinishReason *string         `json:"finish_reason"`
}

type ChatStreamChunk struct {
	ID                string             `json:"id"`
	Object            string             `json:"object"`
	Created           int64              `json:"created"`
	Model             string             `json:"model"`
	Choices           []ChatStreamChoice `json:"choices"`
	ServiceTier       string             `json:"service_tier,omitempty"`
	SystemFingerprint string             `json:"system_fingerprint,omitempty"`
	Usage             *UsageStats        `json:"usage,omitempty"`
}
