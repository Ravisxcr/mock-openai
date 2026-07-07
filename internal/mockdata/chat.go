package mockdata

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"mock-openai/internal/models"
)

const chatFingerprint = "fp_mock000000"

func randomSuffix() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	out := make([]byte, len(b))
	for i, v := range b {
		out[i] = alphabet[int(v)%len(alphabet)]
	}
	return string(out)
}

// estimateTokens is a rough char-based heuristic (~4 chars/token), the
// same rule of thumb OpenAI documents for quick estimates.
func estimateTokens(text string) int {
	if len(text) == 0 {
		return 0
	}
	tokens := len(text) / 4
	if tokens < 1 {
		tokens = 1
	}
	return tokens
}

func replyContent(messages []models.ChatMessage) string {
	if len(messages) == 0 {
		return "This is a mock response from your Go server."
	}
	last := messages[len(messages)-1]
	return fmt.Sprintf("Mock response to your %s message: %s", last.Role, last.Content)
}

// truncateToTokens trims content to roughly maxTokens tokens (using the
// same ~4 chars/token heuristic) and reports whether truncation happened.
func truncateToTokens(content string, maxTokens int) (string, bool) {
	if maxTokens <= 0 {
		return "", true
	}
	limit := maxTokens * 4
	if len(content) <= limit {
		return content, false
	}
	return strings.TrimSpace(content[:limit]), true
}

func buildLogprobs(requested bool, topLogprobs *int, content string) interface{} {
	if !requested {
		return nil
	}
	words := strings.Fields(content)
	tokens := make([]map[string]interface{}, 0, len(words))
	for i, w := range words {
		entry := map[string]interface{}{
			"token":        w,
			"logprob":      -0.01 * float64(i+1),
			"bytes":        []byte(w),
			"top_logprobs": []interface{}{},
		}
		tokens = append(tokens, entry)
	}
	_ = topLogprobs // top_logprobs count intentionally not fanned out further in this mock
	return map[string]interface{}{
		"content": tokens,
		"refusal": nil,
	}
}

func effectiveMaxTokens(req models.ChatRequest) *int {
	if req.MaxCompletionTokens != nil {
		return req.MaxCompletionTokens
	}
	return req.MaxTokens
}

// BuildChatResponse generates a mock CreateChatCompletionResponse for the request.
func BuildChatResponse(req models.ChatRequest) models.ChatResponse {
	n := 1
	if req.N != nil && *req.N > 0 {
		n = *req.N
	}

	promptText := make([]string, 0, len(req.Messages))
	for _, m := range req.Messages {
		promptText = append(promptText, m.Content)
	}
	promptTokens := estimateTokens(strings.Join(promptText, " "))
	maxTokens := effectiveMaxTokens(req)
	base := replyContent(req.Messages)

	choices := make([]models.ChatChoice, n)
	completionTokens := 0
	for i := 0; i < n; i++ {
		content := base
		finish := "stop"
		if maxTokens != nil {
			truncated, hit := truncateToTokens(content, *maxTokens)
			content = truncated
			if hit {
				finish = "length"
			}
		}
		completionTokens += estimateTokens(content)
		choices[i] = models.ChatChoice{
			Index: i,
			Message: models.ChatResponseMessage{
				Role:        "assistant",
				Content:     &content,
				Refusal:     nil,
				Annotations: []interface{}{},
			},
			Logprobs:     buildLogprobs(req.Logprobs, req.TopLogprobs, content),
			FinishReason: finish,
		}
	}

	return models.ChatResponse{
		ID:      "chatcmpl-" + randomSuffix(),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: choices,
		Usage: models.UsageStats{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
		ServiceTier:       "default",
		SystemFingerprint: chatFingerprint,
		RequestMessages:   req.Messages,
	}
}

// BuildStreamChunks turns an already-built ChatResponse back into the
// ordered sequence of chat.completion.chunk events an SSE stream would
// emit: one role-only chunk, then word-by-word content deltas, then a
// finish-reason chunk, per choice, optionally followed by a usage-only
// chunk when the client asked for it via stream_options.include_usage.
func BuildStreamChunks(resp models.ChatResponse, includeUsage bool) []models.ChatStreamChunk {
	var chunks []models.ChatStreamChunk

	base := func() models.ChatStreamChunk {
		return models.ChatStreamChunk{
			ID:                resp.ID,
			Object:            "chat.completion.chunk",
			Created:           resp.Created,
			Model:             resp.Model,
			ServiceTier:       resp.ServiceTier,
			SystemFingerprint: resp.SystemFingerprint,
		}
	}

	for _, choice := range resp.Choices {
		roleChunk := base()
		roleChunk.Choices = []models.ChatStreamChoice{{
			Index:    choice.Index,
			Delta:    models.ChatStreamDelta{Role: "assistant"},
			Logprobs: nil,
		}}
		chunks = append(chunks, roleChunk)

		content := ""
		if choice.Message.Content != nil {
			content = *choice.Message.Content
		}
		words := strings.Fields(content)
		for i, w := range words {
			piece := w
			if i < len(words)-1 {
				piece += " "
			}
			deltaChunk := base()
			deltaChunk.Choices = []models.ChatStreamChoice{{
				Index:    choice.Index,
				Delta:    models.ChatStreamDelta{Content: &piece},
				Logprobs: nil,
			}}
			chunks = append(chunks, deltaChunk)
		}

		finishReason := choice.FinishReason
		finishChunk := base()
		finishChunk.Choices = []models.ChatStreamChoice{{
			Index:        choice.Index,
			Delta:        models.ChatStreamDelta{},
			Logprobs:     nil,
			FinishReason: &finishReason,
		}}
		chunks = append(chunks, finishChunk)
	}

	if includeUsage {
		usageChunk := base()
		usageChunk.Choices = []models.ChatStreamChoice{}
		usage := resp.Usage
		usageChunk.Usage = &usage
		chunks = append(chunks, usageChunk)
	}

	return chunks
}

// GenerateLegacyCompletion returns the non-chat format response (POST /v1/completions).
func GenerateLegacyCompletion(model string) map[string]interface{} {
	return map[string]interface{}{
		"id":      "cmpl-" + randomSuffix(),
		"object":  "text_completion",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []map[string]interface{}{
			{
				"text":          "This is a mock response from the legacy completion API.",
				"index":         0,
				"finish_reason": "stop",
			},
		},
	}
}
