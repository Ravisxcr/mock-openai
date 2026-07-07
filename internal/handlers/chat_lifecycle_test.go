package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"mock-openai/internal/models"
)

func newChatRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/v1/completions", HandleLegacyCompletions)
	r.POST("/v1/chat/completions", HandleChatCompletions)
	r.GET("/v1/chat/completions", HandleListChatCompletions)
	r.GET("/v1/chat/completions/:completion_id", HandleChatCompletion)
	r.POST("/v1/chat/completions/:completion_id", HandleUpdateChatCompletion)
	r.DELETE("/v1/chat/completions/:completion_id", HandleDeleteChatCompletion)
	r.GET("/v1/chat/completions/:completion_id/messages", HandleGetChatMessages)
	return r
}

func createTestCompletion(t *testing.T, router *gin.Engine) models.ChatResponse {
	t.Helper()
	w := postJSON(t, router, "/v1/chat/completions", models.ChatRequest{
		Model:    "gpt-4o",
		Messages: []models.ChatMessage{{Role: "user", Content: models.TextContent("Hello!")}},
	})
	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.ChatResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp
}

func TestHandleLegacyCompletions(t *testing.T) {
	router := newChatRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/completions", bytes.NewBufferString(`{"model":"gpt-3.5-turbo-instruct","prompt":"hi"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "text_completion", body["object"])
	assert.Equal(t, "gpt-3.5-turbo-instruct", body["model"])
}

func TestHandleChatCompletionsMissingModel(t *testing.T) {
	router := newChatRouter()

	w := postJSON(t, router, "/v1/chat/completions", models.ChatRequest{
		Messages: []models.ChatMessage{{Role: "user", Content: models.TextContent("hi")}},
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "model parameter")
}

func TestHandleChatCompletionsMissingMessages(t *testing.T) {
	router := newChatRouter()

	w := postJSON(t, router, "/v1/chat/completions", models.ChatRequest{
		Model: "gpt-4o",
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "'messages' must contain at least one item")
}

func TestHandleChatCompletionsInvalidJSON(t *testing.T) {
	router := newChatRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString("{not-json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleChatCompletionsToolChoiceWithoutTools(t *testing.T) {
	router := newChatRouter()

	req := models.ChatRequest{
		Model:      "gpt-4o",
		Messages:   []models.ChatMessage{{Role: "user", Content: models.TextContent("hi")}},
		ToolChoice: json.RawMessage(`"required"`),
	}
	w := postJSON(t, router, "/v1/chat/completions", req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "'tool_choice' is only allowed when 'tools' are specified")
}

func TestHandleChatCompletionsInvalidToolChoiceValue(t *testing.T) {
	router := newChatRouter()

	req := models.ChatRequest{
		Model:      "gpt-4o",
		Messages:   []models.ChatMessage{{Role: "user", Content: models.TextContent("hi")}},
		Tools:      []models.ChatTool{{Type: "function", Function: models.ChatToolFunction{Name: "noop"}}},
		ToolChoice: json.RawMessage(`123`),
	}
	w := postJSON(t, router, "/v1/chat/completions", req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid value for 'tool_choice'")
}

func TestHandleChatCompletionsStreamingWithUsage(t *testing.T) {
	router := newChatRouter()

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}],"stream":true,"stream_options":{"include_usage":true}}`)
	req, _ := http.NewRequest(http.MethodPost, "/v1/chat/completions", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"usage"`)
	assert.Contains(t, w.Body.String(), "data: [DONE]")
}

func TestHandleUpdateChatCompletionInvalidJSON(t *testing.T) {
	router := newChatRouter()
	created := createTestCompletion(t, router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/chat/completions/"+created.ID, bytes.NewBufferString("{not-json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleListChatCompletions(t *testing.T) {
	router := newChatRouter()
	created := createTestCompletion(t, router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/chat/completions", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var body struct {
		Object string                `json:"object"`
		Data   []models.ChatResponse `json:"data"`
	}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "list", body.Object)
	found := false
	for _, c := range body.Data {
		if c.ID == created.ID {
			found = true
		}
	}
	assert.True(t, found)
}

func TestHandleChatCompletionGetFound(t *testing.T) {
	router := newChatRouter()
	created := createTestCompletion(t, router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/chat/completions/"+created.ID, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var fetched models.ChatResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &fetched))
	assert.Equal(t, created.ID, fetched.ID)
}

func TestHandleChatCompletionGetNotFound(t *testing.T) {
	router := newChatRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/chat/completions/chatcmpl-does-not-exist", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleUpdateChatCompletion(t *testing.T) {
	router := newChatRouter()
	created := createTestCompletion(t, router)

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"metadata":{"env":"test"}}`)
	req, _ := http.NewRequest(http.MethodPost, "/v1/chat/completions/"+created.ID, body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var updated models.ChatResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &updated))
	assert.Equal(t, "test", updated.Metadata["env"])

	// Confirm the update persisted in the store.
	wGet := httptest.NewRecorder()
	reqGet, _ := http.NewRequest(http.MethodGet, "/v1/chat/completions/"+created.ID, nil)
	router.ServeHTTP(wGet, reqGet)
	var fetched models.ChatResponse
	assert.NoError(t, json.Unmarshal(wGet.Body.Bytes(), &fetched))
	assert.Equal(t, "test", fetched.Metadata["env"])
}

func TestHandleUpdateChatCompletionNotFound(t *testing.T) {
	router := newChatRouter()

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"metadata":{"env":"test"}}`)
	req, _ := http.NewRequest(http.MethodPost, "/v1/chat/completions/chatcmpl-does-not-exist", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleDeleteChatCompletion(t *testing.T) {
	router := newChatRouter()
	created := createTestCompletion(t, router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/v1/chat/completions/"+created.ID, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"deleted":true`)

	// Now gone.
	wGet := httptest.NewRecorder()
	reqGet, _ := http.NewRequest(http.MethodGet, "/v1/chat/completions/"+created.ID, nil)
	router.ServeHTTP(wGet, reqGet)
	assert.Equal(t, http.StatusNotFound, wGet.Code)
}

func TestHandleDeleteChatCompletionNotFound(t *testing.T) {
	router := newChatRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/v1/chat/completions/chatcmpl-does-not-exist", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleGetChatMessagesNotFound(t *testing.T) {
	router := newChatRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/chat/completions/chatcmpl-does-not-exist/messages", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleChatCompletionsStreaming(t *testing.T) {
	router := newChatRouter()

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}],"stream":true}`)
	req, _ := http.NewRequest(http.MethodPost, "/v1/chat/completions", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "data: [DONE]")
}
