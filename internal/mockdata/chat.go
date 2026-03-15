package mockdata

import (
	"time"

	"github.com/gin-gonic/gin"
	"mock-openai/internal/models"
)

// GenerateMockChatResponse creates a standard OpenAI-style response object.
func GenerateMockChatResponse(modelName string, content string) models.ChatResponse {
	return models.ChatResponse{
		ID:      "chatcmpl-mock-" + time.Now().Format("20060102150405"), // Dynamic ID based on timestamp
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   modelName,
		Choices: []models.ChatChoice{
			{
				Index: 0,
				Message: models.ChatMessage{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: "stop",
			},
		},
		Usage: models.UsageStats{
			PromptTokens:     15,
			CompletionTokens: 25,
			TotalTokens:      40,
		},
	}
}

// GenerateLegacyCompletion returns the non-chat format response
func GenerateLegacyCompletion(model string) gin.H {
	return gin.H{
		"id":      "cmpl-mock-" + time.Now().Format("20060102150405"),
		"object":  "text_completion",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []gin.H{
			{
				"text":          "This is a mock response from the legacy completion API.",
				"index":         0,
				"finish_reason": "stop",
			},
		},
	}
}

// GenerateMockMessageList returns a history of messages
func GenerateMockMessageList() []models.ChatMessage {
	return []models.ChatMessage{
		{Role: "user", Content: "Hello, can you help me?"},
		{Role: "assistant", Content: "Of course! I am your mock AI assistant."},
	}
}
