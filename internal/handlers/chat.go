package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"mock-openai/internal/models" // Update this path
	"mock-openai/internal/mockdata" // Update this path
)

// HandleChatCompletions mimics the OpenAI Chat API logic.
func HandleChatCompletions(c *gin.Context) {
	var req models.ChatRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Simple logic: Respond based on the last message sent
	content := "This is a mock response from your Go server."
	if len(req.Messages) > 0 {
		lastMsg := req.Messages[len(req.Messages)-1].Content
		content = "Mock Assistant: I received your message: " + lastMsg
	}

	res := mockdata.GenerateMockChatResponse(req.Model, content)


	c.JSON(http.StatusOK, res)
}

func HandleModels(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"data": []gin.H{
			{"id": "gpt-3.5-turbo", "object": "model", "created": 1677610602, "owned_by": "openai"},
			{"id": "gpt-4", "object": "model", "created": 1680000000, "owned_by": "openai"},
			{"id": "gpt-4-32k", "object": "model", "created": 1680000000, "owned_by": "openai"},
		},
	})
}

func HandleModel(c *gin.Context) {
	modelID := c.Param("model_id")
	// For simplicity, we return the same details for any model ID
	c.JSON(http.StatusOK, gin.H{
		"id":         modelID,
		"object":     "model",
		"created":    1677610602,
		"owned_by":   "openai",
		"description": "This is a mock model description for " + modelID,
	})
}


// HandleChat processes incoming chat completion requests.
func HandleChat(c *gin.Context) {
	var req models.ChatRequest

	// Bind the incoming JSON to our struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Logic: We respond with a mock message that repeats the last thing the user said
	content := "I am a mock response."
	if len(req.Messages) > 0 {
		content = "Mock Assistant: You said: " + req.Messages[len(req.Messages)-1].Content
	}

	// Build the OpenAI-compatible response
	resp := models.ChatResponse{
		ID:      "chatcmpl-mock-123",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
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
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}

	// Send the JSON response with a 200 OK status
	c.JSON(http.StatusOK, resp)
}

func HandleChatCompletion(c *gin.Context) {

	completion_id := c.Param("completion_id")

	// Logic: We respond with a mock message that repeats the last thing the user said
	content := "I am a mock response."

	// Build the OpenAI-compatible response
	resp := models.ChatResponse{
		ID:      completion_id,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "gpt-4o",
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
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}

	// Send the JSON response with a 200 OK status
	c.JSON(http.StatusOK, resp)
}











// POST /v1/completions
func HandleLegacyCompletions(c *gin.Context) {
	// Typically takes a 'prompt' string instead of 'messages' array
	c.JSON(http.StatusOK, mockdata.GenerateLegacyCompletion("gpt-3.5-turbo-instruct"))
}

// GET /v1/chat/completions
func HandleListChatCompletions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data": []models.ChatResponse{
			mockdata.GenerateMockChatResponse("gpt-4o", "Recent conversation A"),
			mockdata.GenerateMockChatResponse("gpt-4o", "Recent conversation B"),
		},
	})
}

// POST /v1/chat/completions/:completion_id (Update)
func HandleUpdateChatCompletion(c *gin.Context) {
	id := c.Param("completion_id")
	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"object":  "chat.completion",
		"updated": true,
		"status":  "success",
	})
}

// DELETE /v1/chat/completions/:completion_id
func HandleDeleteChatCompletion(c *gin.Context) {
	id := c.Param("completion_id")
	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"object":  "chat.completion",
		"deleted": true,
	})
}

// GET /v1/chat/completions/:completion_id/messages
func HandleGetChatMessages(c *gin.Context) {
	id := c.Param("completion_id")
	c.JSON(http.StatusOK, gin.H{
		"object":        "list",
		"completion_id": id,
		"data":          mockdata.GenerateMockMessageList(),
	})
}