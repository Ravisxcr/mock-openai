package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"mock-openai/internal/apierror"
	"mock-openai/internal/mockdata"
	"mock-openai/internal/models"
	"mock-openai/internal/store"
)

// POST /v1/chat/completions
func HandleChatCompletions(c *gin.Context) {
	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidRequest(c, "Invalid request body: "+err.Error(), "")
		return
	}
	if req.Model == "" {
		apierror.InvalidRequest(c, "you must provide a model parameter", "model")
		return
	}
	if len(req.Messages) == 0 {
		apierror.InvalidRequest(c, "'messages' must contain at least one item", "messages")
		return
	}

	resp := mockdata.BuildChatResponse(req)
	store.SaveCompletion(resp)

	if !req.Stream {
		c.JSON(http.StatusOK, resp)
		return
	}

	streamChatCompletion(c, resp, req)
}

func streamChatCompletion(c *gin.Context, resp models.ChatResponse, req models.ChatRequest) {
	includeUsage := req.StreamOptions != nil && req.StreamOptions.IncludeUsage
	chunks := mockdata.BuildStreamChunks(resp, includeUsage)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Status(http.StatusOK)

	flusher, canFlush := c.Writer.(http.Flusher)
	for _, chunk := range chunks {
		data, _ := json.Marshal(chunk)
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
		if canFlush {
			flusher.Flush()
		}
		time.Sleep(20 * time.Millisecond)
	}
	fmt.Fprint(c.Writer, "data: [DONE]\n\n")
	if canFlush {
		flusher.Flush()
	}
}

// GET /v1/chat/completions
func HandleListChatCompletions(c *gin.Context) {
	completions := store.ListCompletions()
	ids := make([]string, len(completions))
	for i, cpl := range completions {
		ids[i] = cpl.ID
	}
	firstID, lastID := "", ""
	if len(ids) > 0 {
		firstID, lastID = ids[0], ids[len(ids)-1]
	}
	c.JSON(http.StatusOK, gin.H{
		"object":   "list",
		"data":     completions,
		"first_id": firstID,
		"last_id":  lastID,
		"has_more": false,
	})
}

// GET /v1/chat/completions/:completion_id
func HandleChatCompletion(c *gin.Context) {
	id := c.Param("completion_id")
	resp, ok := store.GetCompletion(id)
	if !ok {
		apierror.NotFound(c, fmt.Sprintf("No chat completion found with id '%s'.", id), "completion_not_found")
		return
	}
	c.JSON(http.StatusOK, resp)
}

// POST /v1/chat/completions/:completion_id (metadata-only update, per spec)
func HandleUpdateChatCompletion(c *gin.Context) {
	id := c.Param("completion_id")
	resp, ok := store.GetCompletion(id)
	if !ok {
		apierror.NotFound(c, fmt.Sprintf("No chat completion found with id '%s'.", id), "completion_not_found")
		return
	}

	var body struct {
		Metadata map[string]string `json:"metadata"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		apierror.InvalidRequest(c, "Invalid request body: "+err.Error(), "")
		return
	}
	resp.Metadata = body.Metadata
	store.SaveCompletion(resp)
	c.JSON(http.StatusOK, resp)
}

// DELETE /v1/chat/completions/:completion_id
func HandleDeleteChatCompletion(c *gin.Context) {
	id := c.Param("completion_id")
	if !store.DeleteCompletion(id) {
		apierror.NotFound(c, fmt.Sprintf("No chat completion found with id '%s'.", id), "completion_not_found")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"object":  "chat.completion.deleted",
		"deleted": true,
	})
}

// GET /v1/chat/completions/:completion_id/messages
func HandleGetChatMessages(c *gin.Context) {
	id := c.Param("completion_id")
	resp, ok := store.GetCompletion(id)
	if !ok {
		apierror.NotFound(c, fmt.Sprintf("No chat completion found with id '%s'.", id), "completion_not_found")
		return
	}

	data := make([]gin.H, 0, len(resp.RequestMessages)+1)
	for _, m := range resp.RequestMessages {
		data = append(data, gin.H{"role": m.Role, "content": m.Content})
	}
	if len(resp.Choices) > 0 {
		data = append(data, gin.H{"role": "assistant", "content": resp.Choices[0].Message.Content})
	}

	c.JSON(http.StatusOK, gin.H{
		"object":   "list",
		"data":     data,
		"has_more": false,
	})
}

// POST /v1/completions (legacy)
func HandleLegacyCompletions(c *gin.Context) {
	c.JSON(http.StatusOK, mockdata.GenerateLegacyCompletion("gpt-3.5-turbo-instruct"))
}
