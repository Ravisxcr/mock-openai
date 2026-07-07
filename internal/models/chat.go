package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// ChatMessage represents a single message in a conversation.
type ChatMessage struct {
	Role       string             `json:"role"`
	Content    ChatMessageContent `json:"content"`
	Name       string             `json:"name,omitempty"`
	ToolCalls  []ToolCall         `json:"tool_calls,omitempty"`   // assistant messages replaying tool-call history
	ToolCallID string             `json:"tool_call_id,omitempty"` // tool-role messages: which call this responds to
}

// ChatImageURL mirrors the `image_url` object inside
// ChatCompletionRequestMessageContentPartImage.
type ChatImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// ChatInputAudio mirrors the `input_audio` object inside
// ChatCompletionRequestMessageContentPartAudio.
type ChatInputAudio struct {
	Data   string `json:"data"`
	Format string `json:"format"`
}

// ChatFilePart mirrors the `file` object inside
// ChatCompletionRequestMessageContentPartFile.
type ChatFilePart struct {
	Filename string `json:"filename,omitempty"`
	FileData string `json:"file_data,omitempty"`
	FileID   string `json:"file_id,omitempty"`
}

// ChatContentPart mirrors the union of ChatCompletionRequestMessageContentPart*
// schemas (text/image_url/input_audio/file/refusal) in openai.yml, discriminated
// by Type. This mock does not enforce which part types are valid for which
// message role (real API: system/tool allow text-only, assistant allows
// text/refusal, user allows text/image_url/input_audio/file) — see memory.md.
type ChatContentPart struct {
	Type       string          `json:"type"`
	Text       string          `json:"text,omitempty"`
	ImageURL   *ChatImageURL   `json:"image_url,omitempty"`
	InputAudio *ChatInputAudio `json:"input_audio,omitempty"`
	File       *ChatFilePart   `json:"file,omitempty"`
	Refusal    string          `json:"refusal,omitempty"`
}

// ChatMessageContent represents the oneOf(string | array-of-parts) `content`
// field shared by the request message schemas, plus the `null` case allowed
// for assistant messages that only carry tool_calls. Exactly one of Text or
// Parts is set; both nil means the JSON value was null (or absent).
type ChatMessageContent struct {
	Text  *string
	Parts []ChatContentPart
}

// TextContent builds a plain string ChatMessageContent, for constructing
// requests/messages in Go code (tests, etc.) without going through JSON.
func TextContent(s string) ChatMessageContent {
	return ChatMessageContent{Text: &s}
}

func (c *ChatMessageContent) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		*c = ChatMessageContent{}
		return nil
	}
	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*c = ChatMessageContent{Text: &s}
		return nil
	}
	var parts []ChatContentPart
	if err := json.Unmarshal(data, &parts); err != nil {
		return fmt.Errorf("content must be a string, an array of content parts, or null: %w", err)
	}
	*c = ChatMessageContent{Parts: parts}
	return nil
}

func (c ChatMessageContent) MarshalJSON() ([]byte, error) {
	if c.Parts != nil {
		return json.Marshal(c.Parts)
	}
	if c.Text != nil {
		return json.Marshal(*c.Text)
	}
	return []byte("null"), nil
}

// maxInlineURLLen truncates long content-part payloads (e.g. base64 data
// URIs in image_url.url) before they're folded into the mock's flattened
// text, so a single multimodal message doesn't blow up token/echo output.
const maxInlineURLLen = 60

func truncateForDisplay(s string) string {
	if len(s) <= maxInlineURLLen {
		return s
	}
	return s[:maxInlineURLLen] + "...(truncated)"
}

// Flatten collapses content (string or content parts) down to a single
// string for the mock's text-based logic (reply generation, token
// estimation). Non-text parts are rendered as human-readable placeholders
// rather than being dropped, so their presence is still observable.
func (c ChatMessageContent) Flatten() string {
	if c.Text != nil {
		return *c.Text
	}
	pieces := make([]string, 0, len(c.Parts))
	for _, p := range c.Parts {
		switch p.Type {
		case "text":
			pieces = append(pieces, p.Text)
		case "image_url":
			if p.ImageURL != nil {
				pieces = append(pieces, fmt.Sprintf("[image: %s]", truncateForDisplay(p.ImageURL.URL)))
			}
		case "input_audio":
			if p.InputAudio != nil {
				pieces = append(pieces, fmt.Sprintf("[audio input, format=%s]", p.InputAudio.Format))
			}
		case "file":
			if p.File != nil {
				name := p.File.Filename
				if name == "" {
					name = p.File.FileID
				}
				pieces = append(pieces, fmt.Sprintf("[file: %s]", name))
			}
		case "refusal":
			pieces = append(pieces, p.Refusal)
		}
	}
	return strings.Join(pieces, " ")
}

// ChatToolFunction mirrors FunctionObject in openai.yml.
type ChatToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Strict      *bool                  `json:"strict,omitempty"`
}

// ChatTool mirrors ChatCompletionTool in openai.yml.
type ChatTool struct {
	Type     string           `json:"type"`
	Function ChatToolFunction `json:"function"`
}

// ToolCallFunction mirrors the `function` object inside ChatCompletionMessageToolCall.
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolCall mirrors ChatCompletionMessageToolCall (Index omitted) and
// ChatCompletionMessageToolCallChunk (Index set) — the two schemas share this
// shape everywhere except streaming chunks require an `index`.
type ToolCall struct {
	Index    *int             `json:"index,omitempty"`
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
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
	Tools               []ChatTool         `json:"tools,omitempty"`
	ToolChoice          json.RawMessage    `json:"tool_choice,omitempty"`
	ParallelToolCalls   *bool              `json:"parallel_tool_calls,omitempty"`
}

// ChatResponseMessage is the assistant message returned in a completion choice.
type ChatResponseMessage struct {
	Role        string        `json:"role"`
	Content     *string       `json:"content"`
	Refusal     *string       `json:"refusal"`
	Annotations []interface{} `json:"annotations"`
	ToolCalls   []ToolCall    `json:"tool_calls,omitempty"`
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
	Role      string     `json:"role,omitempty"`
	Content   *string    `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
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
