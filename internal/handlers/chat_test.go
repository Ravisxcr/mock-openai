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
	"mock-openai/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestHandleChatConcurrency(t *testing.T) {
	// 1. Setup Gin in Test Mode
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/chat/completions", HandleChat)

	// 2. Configuration for the test
	const concurrentRequests = 100
	var wg sync.WaitGroup
	wg.Add(concurrentRequests)

	// 3. Prepare the request body
	requestBody, _ := json.Marshal(models.ChatRequest{
		Model: "gpt-4o",
		Messages: []models.ChatMessage{
			{Role: "user", Content: "Hello Concurrency!"},
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

