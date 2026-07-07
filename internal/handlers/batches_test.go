package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"mock-openai/internal/models"
)

func newBatchesRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/v1/files", HandleCreateFile)
	r.POST("/v1/batches", HandleCreateBatch)
	r.GET("/v1/batches", HandleListBatches)
	r.GET("/v1/batches/:batch_id", HandleGetBatch)
	r.POST("/v1/batches/:batch_id/cancel", HandleCancelBatch)
	return r
}

const batchInputJSONL = `{"custom_id":"req-1","method":"POST","url":"/v1/chat/completions","body":{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}]}}
`

func TestHandleCreateBatchSuccess(t *testing.T) {
	router := newBatchesRouter()
	inputFile := createTestFile(t, router, "batch_input.jsonl", "batch", []byte(batchInputJSONL))

	w := postJSON(t, router, "/v1/batches", models.CreateBatchRequest{
		InputFileID:      inputFile.ID,
		Endpoint:         "/v1/chat/completions",
		CompletionWindow: "24h",
	})
	assert.Equal(t, http.StatusOK, w.Code)

	var batch models.Batch
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &batch))
	assert.Equal(t, "batch", batch.Object)
	assert.Equal(t, inputFile.ID, batch.InputFileID)
	assert.Equal(t, "/v1/chat/completions", batch.Endpoint)
	assert.NotEmpty(t, batch.ID)
	assert.NotNil(t, batch.OutputFileID)
	assert.Equal(t, 1, batch.RequestCounts.Total)
	assert.Equal(t, 1, batch.RequestCounts.Completed)
}

func TestHandleCreateBatchValidation(t *testing.T) {
	router := newBatchesRouter()
	inputFile := createTestFile(t, router, "batch_input.jsonl", "batch", []byte(batchInputJSONL))

	cases := []struct {
		name string
		req  models.CreateBatchRequest
		want string
	}{
		{"missing input_file_id", models.CreateBatchRequest{Endpoint: "/v1/chat/completions", CompletionWindow: "24h"}, "input_file_id parameter"},
		{"missing endpoint", models.CreateBatchRequest{InputFileID: inputFile.ID, CompletionWindow: "24h"}, "endpoint parameter"},
		{"missing completion_window", models.CreateBatchRequest{InputFileID: inputFile.ID, Endpoint: "/v1/chat/completions"}, "completion_window parameter"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := postJSON(t, router, "/v1/batches", tc.req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), tc.want)
		})
	}
}

func TestHandleCreateBatchUnknownInputFile(t *testing.T) {
	router := newBatchesRouter()

	w := postJSON(t, router, "/v1/batches", models.CreateBatchRequest{
		InputFileID:      "file-does-not-exist",
		Endpoint:         "/v1/chat/completions",
		CompletionWindow: "24h",
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "No such file")
}

func TestHandleListBatches(t *testing.T) {
	router := newBatchesRouter()
	inputFile := createTestFile(t, router, "batch_input.jsonl", "batch", []byte(batchInputJSONL))
	wCreate := postJSON(t, router, "/v1/batches", models.CreateBatchRequest{
		InputFileID:      inputFile.ID,
		Endpoint:         "/v1/chat/completions",
		CompletionWindow: "24h",
	})
	var created models.Batch
	assert.NoError(t, json.Unmarshal(wCreate.Body.Bytes(), &created))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/batches", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var body struct {
		Object string         `json:"object"`
		Data   []models.Batch `json:"data"`
	}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "list", body.Object)
	found := false
	for _, b := range body.Data {
		if b.ID == created.ID {
			found = true
		}
	}
	assert.True(t, found)
}

func TestHandleGetBatchFound(t *testing.T) {
	router := newBatchesRouter()
	inputFile := createTestFile(t, router, "batch_input.jsonl", "batch", []byte(batchInputJSONL))
	wCreate := postJSON(t, router, "/v1/batches", models.CreateBatchRequest{
		InputFileID:      inputFile.ID,
		Endpoint:         "/v1/chat/completions",
		CompletionWindow: "24h",
	})
	var created models.Batch
	assert.NoError(t, json.Unmarshal(wCreate.Body.Bytes(), &created))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/batches/"+created.ID, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var fetched models.Batch
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &fetched))
	assert.Equal(t, created.ID, fetched.ID)
}

func TestHandleGetBatchNotFound(t *testing.T) {
	router := newBatchesRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/batches/batch_does_not_exist", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleCancelBatch(t *testing.T) {
	router := newBatchesRouter()
	inputFile := createTestFile(t, router, "batch_input.jsonl", "batch", []byte(batchInputJSONL))
	wCreate := postJSON(t, router, "/v1/batches", models.CreateBatchRequest{
		InputFileID:      inputFile.ID,
		Endpoint:         "/v1/chat/completions",
		CompletionWindow: "24h",
	})
	var created models.Batch
	assert.NoError(t, json.Unmarshal(wCreate.Body.Bytes(), &created))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/batches/"+created.ID+"/cancel", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var cancelled models.Batch
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &cancelled))
	assert.Equal(t, created.ID, cancelled.ID)
}

func TestHandleCancelBatchNotFound(t *testing.T) {
	router := newBatchesRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/batches/batch_does_not_exist/cancel", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
