package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	// "github.com/youruser/mock-openai/internal/chat" // Update this path
	"github.com/stretchr/testify/assert"
	"mock-openai/internal/models"
)

func TestHandleChatConcurrency(t *testing.T) {
	// 1. Setup Gin in Test Mode
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/chat/completions", HandleChatCompletions)

	// 2. Configuration for the test
	const concurrentRequests = 100
	var wg sync.WaitGroup
	wg.Add(concurrentRequests)

	// 3. Prepare the request body
	requestBody, _ := json.Marshal(models.ChatRequest{
		Model: "gpt-4o",
		Messages: []models.ChatMessage{
			{Role: "user", Content: models.TextContent("Hello Concurrency!")},
		},
	})

	// 4. Fire requests concurrently
	for i := 0; i < concurrentRequests; i++ {
		go func() {
			defer wg.Done()

			// Create a response recorder for each goroutine
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Perform the request
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, http.StatusOK, w.Code)

			var response models.ChatResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "chat.completion", response.Object)
		}()
	}

	// 5. Wait for all goroutines to finish
	wg.Wait()
}

func TestHandleChatMultimodalContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/chat/completions", HandleChatCompletions)
	router.GET("/v1/chat/completions/:completion_id/messages", HandleGetChatMessages)

	rawBody := []byte(`{
		"model": "gpt-4o",
		"messages": [
			{
				"role": "user",
				"content": [
					{"type": "text", "text": "What's in this image?"},
					{"type": "image_url", "image_url": {"url": "https://example.com/cat.png", "detail": "high"}}
				]
			}
		]
	}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(rawBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.ChatResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	if assert.Len(t, resp.Choices, 1) {
		assert.NotNil(t, resp.Choices[0].Message.Content)
		// Reply text is derived from the flattened content, so the image
		// placeholder should surface in the mock's response.
		assert.Contains(t, *resp.Choices[0].Message.Content, "image")
		assert.Contains(t, *resp.Choices[0].Message.Content, "example.com/cat.png")
	}

	// GET .../messages should echo the original array-of-parts shape back,
	// not collapse it to a flattened string.
	wMsgs := httptest.NewRecorder()
	reqMsgs, _ := http.NewRequest("GET", "/v1/chat/completions/"+resp.ID+"/messages", nil)
	router.ServeHTTP(wMsgs, reqMsgs)
	assert.Equal(t, http.StatusOK, wMsgs.Code)

	var msgsResp struct {
		Data []struct {
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
		} `json:"data"`
	}
	assert.NoError(t, json.Unmarshal(wMsgs.Body.Bytes(), &msgsResp))
	if assert.Len(t, msgsResp.Data, 2) {
		var parts []models.ChatContentPart
		assert.NoError(t, json.Unmarshal(msgsResp.Data[0].Content, &parts))
		if assert.Len(t, parts, 2) {
			assert.Equal(t, "text", parts[0].Type)
			assert.Equal(t, "image_url", parts[1].Type)
			assert.Equal(t, "https://example.com/cat.png", parts[1].ImageURL.URL)
		}
	}
}

func TestHandleChatToolCallRoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/chat/completions", HandleChatCompletions)

	weatherTool := models.ChatTool{
		Type: "function",
		Function: models.ChatToolFunction{
			Name:        "get_current_weather",
			Description: "Get the current weather in a location",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"location": map[string]interface{}{"type": "string"},
					"unit":     map[string]interface{}{"type": "string", "enum": []interface{}{"celsius", "fahrenheit"}},
				},
				"required": []interface{}{"location"},
			},
		},
	}

	post := func(t *testing.T, req models.ChatRequest) (*httptest.ResponseRecorder, models.ChatResponse) {
		t.Helper()
		body, err := json.Marshal(req)
		assert.NoError(t, err)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, httpReq)
		var resp models.ChatResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		return w, resp
	}

	// First turn: tools provided, default tool_choice -> expect a tool call.
	firstReq := models.ChatRequest{
		Model:    "gpt-4o",
		Messages: []models.ChatMessage{{Role: "user", Content: models.TextContent("What's the weather in Boston?")}},
		Tools:    []models.ChatTool{weatherTool},
	}
	w, resp := post(t, firstReq)
	assert.Equal(t, http.StatusOK, w.Code)
	if assert.Len(t, resp.Choices, 1) {
		choice := resp.Choices[0]
		assert.Equal(t, "tool_calls", choice.FinishReason)
		assert.Nil(t, choice.Message.Content)
		if assert.Len(t, choice.Message.ToolCalls, 1) {
			tc := choice.Message.ToolCalls[0]
			assert.Equal(t, "function", tc.Type)
			assert.Equal(t, "get_current_weather", tc.Function.Name)
			var args map[string]interface{}
			assert.NoError(t, json.Unmarshal([]byte(tc.Function.Arguments), &args))
			assert.Contains(t, args, "location")

			// Second turn: client appends the assistant tool-call message and a
			// tool-role result -> expect a normal text finish, no more tool calls.
			secondReq := models.ChatRequest{
				Model: "gpt-4o",
				Messages: []models.ChatMessage{
					{Role: "user", Content: models.TextContent("What's the weather in Boston?")},
					{Role: "assistant", ToolCalls: []models.ToolCall{tc}},
					{Role: "tool", ToolCallID: tc.ID, Content: models.TextContent("{\"temp\":72}")},
				},
				Tools: []models.ChatTool{weatherTool},
			}
			w2, resp2 := post(t, secondReq)
			assert.Equal(t, http.StatusOK, w2.Code)
			if assert.Len(t, resp2.Choices, 1) {
				assert.Equal(t, "stop", resp2.Choices[0].FinishReason)
				assert.Empty(t, resp2.Choices[0].Message.ToolCalls)
			}
		}
	}

	// tool_choice: "none" -> never call, even with tools present.
	noneReq := firstReq
	noneReq.ToolChoice = json.RawMessage(`"none"`)
	_, resp3 := post(t, noneReq)
	if assert.Len(t, resp3.Choices, 1) {
		assert.Equal(t, "stop", resp3.Choices[0].FinishReason)
		assert.Empty(t, resp3.Choices[0].Message.ToolCalls)
	}

	// tool_choice naming an unknown function -> 400 invalid_request_error.
	badReq := firstReq
	badReq.ToolChoice = json.RawMessage(`{"type":"function","function":{"name":"nonexistent"}}`)
	wBad, _ := post(t, badReq)
	assert.Equal(t, http.StatusBadRequest, wBad.Code)
}
