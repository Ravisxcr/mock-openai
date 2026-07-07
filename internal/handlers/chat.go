package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"mock-openai/internal/errs"
	"mock-openai/internal/mockdata"
	"mock-openai/internal/models"
	"mock-openai/internal/store"
)

// HandleChatCompletions creates a model response for the given chat conversation.
//
// @Summary		Create chat completion
// @Description	Creates a model response for the given chat conversation. Supports streaming (SSE), tool/function calling, and multimodal content parts.
// @Tags			Chat
// @Accept			json
// @Produce		json
// @Param			request	body		models.ChatRequest	true	"Chat completion request"
// @Success		200		{object}	models.ChatResponse
// @Failure		400		{object}	errs.Response
// @Failure		401		{object}	errs.Response
// @Security		BearerAuth
// @Router			/chat/completions [post]
func HandleChatCompletions(c *gin.Context) {
	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errs.InvalidRequest(c, "Invalid request body: "+err.Error(), "")
		return
	}
	if req.Model == "" {
		errs.InvalidRequest(c, "you must provide a model parameter", "model")
		return
	}
	if len(req.Messages) == 0 {
		errs.InvalidRequest(c, "'messages' must contain at least one item", "messages")
		return
	}
	if len(req.ToolChoice) > 0 || len(req.Tools) > 0 {
		mode, functionName, err := mockdata.ParseToolChoice(req.ToolChoice, len(req.Tools) > 0)
		if err != nil {
			errs.InvalidRequest(c, "Invalid value for 'tool_choice': "+err.Error(), "tool_choice")
			return
		}
		if len(req.ToolChoice) > 0 && len(req.Tools) == 0 && mode != "none" {
			errs.InvalidRequest(c, "'tool_choice' is only allowed when 'tools' are specified.", "tool_choice")
			return
		}
		if mode == "function" {
			found := false
			for _, t := range req.Tools {
				if t.Function.Name == functionName {
					found = true
					break
				}
			}
			if !found {
				errs.InvalidRequest(c, fmt.Sprintf("Invalid value for 'tool_choice': function '%s' does not exist. Please check your 'tools' array.", functionName), "tool_choice")
				return
			}
		}
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

// HandleListChatCompletions lists stored chat completions.
//
// @Summary		List chat completions
// @Description	Lists chat completions previously created and stored by this mock server.
// @Tags			Chat
// @Produce		json
// @Success		200	{object}	map[string]interface{}
// @Failure		401	{object}	errs.Response
// @Security		BearerAuth
// @Router			/chat/completions [get]
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

// HandleChatCompletion retrieves a stored chat completion by id.
//
// @Summary		Get chat completion
// @Description	Retrieves a stored chat completion by id.
// @Tags			Chat
// @Produce		json
// @Param			completion_id	path		string	true	"Chat completion ID"
// @Success		200				{object}	models.ChatResponse
// @Failure		404				{object}	errs.Response
// @Security		BearerAuth
// @Router			/chat/completions/{completion_id} [get]
func HandleChatCompletion(c *gin.Context) {
	id := c.Param("completion_id")
	resp, ok := store.GetCompletion(id)
	if !ok {
		errs.NotFound(c, fmt.Sprintf("No chat completion found with id '%s'.", id), "completion_not_found")
		return
	}
	c.JSON(http.StatusOK, resp)
}

// HandleUpdateChatCompletion updates the metadata of a stored chat completion.
//
// @Summary		Update chat completion
// @Description	Modifies metadata on a stored chat completion. Only the `metadata` field can be updated, per the real API.
// @Tags			Chat
// @Accept			json
// @Produce		json
// @Param			completion_id	path		string					true	"Chat completion ID"
// @Param			request			body		object{metadata=map[string]string}	true	"Metadata update"
// @Success		200				{object}	models.ChatResponse
// @Failure		400				{object}	errs.Response
// @Failure		404				{object}	errs.Response
// @Security		BearerAuth
// @Router			/chat/completions/{completion_id} [post]
func HandleUpdateChatCompletion(c *gin.Context) {
	id := c.Param("completion_id")
	resp, ok := store.GetCompletion(id)
	if !ok {
		errs.NotFound(c, fmt.Sprintf("No chat completion found with id '%s'.", id), "completion_not_found")
		return
	}

	var body struct {
		Metadata map[string]string `json:"metadata"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		errs.InvalidRequest(c, "Invalid request body: "+err.Error(), "")
		return
	}
	resp.Metadata = body.Metadata
	store.SaveCompletion(resp)
	c.JSON(http.StatusOK, resp)
}

// HandleDeleteChatCompletion deletes a stored chat completion.
//
// @Summary		Delete chat completion
// @Description	Deletes a stored chat completion by id.
// @Tags			Chat
// @Produce		json
// @Param			completion_id	path		string	true	"Chat completion ID"
// @Success		200				{object}	object{id=string,object=string,deleted=bool}
// @Failure		404				{object}	errs.Response
// @Security		BearerAuth
// @Router			/chat/completions/{completion_id} [delete]
func HandleDeleteChatCompletion(c *gin.Context) {
	id := c.Param("completion_id")
	if !store.DeleteCompletion(id) {
		errs.NotFound(c, fmt.Sprintf("No chat completion found with id '%s'.", id), "completion_not_found")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"object":  "chat.completion.deleted",
		"deleted": true,
	})
}

// HandleGetChatMessages lists the messages of a stored chat completion.
//
// @Summary		List chat completion messages
// @Description	Lists the request messages plus the generated assistant reply for a stored chat completion.
// @Tags			Chat
// @Produce		json
// @Param			completion_id	path		string	true	"Chat completion ID"
// @Success		200				{object}	map[string]interface{}
// @Failure		404				{object}	errs.Response
// @Security		BearerAuth
// @Router			/chat/completions/{completion_id}/messages [get]
func HandleGetChatMessages(c *gin.Context) {
	id := c.Param("completion_id")
	resp, ok := store.GetCompletion(id)
	if !ok {
		errs.NotFound(c, fmt.Sprintf("No chat completion found with id '%s'.", id), "completion_not_found")
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

// HandleLegacyCompletions creates a legacy (non-chat) text completion.
//
// @Summary		Create legacy completion
// @Description	Creates a completion for the deprecated /v1/completions endpoint. Always returns a fixed mock response for model "gpt-3.5-turbo-instruct".
// @Tags			Completions
// @Produce		json
// @Success		200	{object}	map[string]interface{}
// @Failure		401	{object}	errs.Response
// @Security		BearerAuth
// @Router			/completions [post]
func HandleLegacyCompletions(c *gin.Context) {
	c.JSON(http.StatusOK, mockdata.GenerateLegacyCompletion("gpt-3.5-turbo-instruct"))
}
