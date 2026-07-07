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

func newEmbeddingsRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/v1/embeddings", HandleEmbeddings)
	return r
}

func postJSON(t *testing.T, router *gin.Engine, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	assert.NoError(t, json.NewEncoder(&buf).Encode(body))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	return w
}

func TestHandleEmbeddingsStringInput(t *testing.T) {
	router := newEmbeddingsRouter()

	w := postJSON(t, router, "/v1/embeddings", models.EmbeddingRequest{
		Model: "text-embedding-3-small",
		Input: "hello world",
	})
	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.EmbeddingResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "list", resp.Object)
	assert.Equal(t, "text-embedding-3-small", resp.Model)
	if assert.Len(t, resp.Data, 1) {
		assert.Equal(t, "embedding", resp.Data[0].Object)
		assert.Equal(t, 0, resp.Data[0].Index)
	}
	assert.Greater(t, resp.Usage.TotalTokens, 0)
}

func TestHandleEmbeddingsArrayInput(t *testing.T) {
	router := newEmbeddingsRouter()

	w := postJSON(t, router, "/v1/embeddings", models.EmbeddingRequest{
		Model: "text-embedding-3-small",
		Input: []interface{}{"first", "second", "third"},
	})
	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.EmbeddingResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 3)
}

func TestHandleEmbeddingsBase64Encoding(t *testing.T) {
	router := newEmbeddingsRouter()

	w := postJSON(t, router, "/v1/embeddings", models.EmbeddingRequest{
		Model:          "text-embedding-3-small",
		Input:          "hello",
		EncodingFormat: "base64",
	})
	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.EmbeddingResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	if assert.Len(t, resp.Data, 1) {
		_, isString := resp.Data[0].Embedding.(string)
		assert.True(t, isString, "expected base64-encoded embedding to be a string")
	}
}

func TestHandleEmbeddingsMissingModel(t *testing.T) {
	router := newEmbeddingsRouter()

	w := postJSON(t, router, "/v1/embeddings", models.EmbeddingRequest{
		Input: "hello world",
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "model parameter")
}

func TestHandleEmbeddingsMissingInput(t *testing.T) {
	router := newEmbeddingsRouter()

	w := postJSON(t, router, "/v1/embeddings", models.EmbeddingRequest{
		Model: "text-embedding-3-small",
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "input parameter")
}

func TestHandleEmbeddingsInvalidJSON(t *testing.T) {
	router := newEmbeddingsRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewBufferString("{not-json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
