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

func newFilesRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/v1/files", HandleCreateFile)
	r.GET("/v1/files", HandleListFiles)
	r.GET("/v1/files/:file_id", HandleGetFile)
	r.DELETE("/v1/files/:file_id", HandleDeleteFile)
	r.GET("/v1/files/:file_id/content", HandleDownloadFileContent)
	return r
}

func createTestFile(t *testing.T, router *gin.Engine, filename, purpose string, content []byte) models.FileObject {
	t.Helper()
	req := newMultipartRequest(t, "/v1/files", map[string]string{"purpose": purpose}, "file", filename, content)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var obj models.FileObject
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &obj))
	return obj
}

func TestHandleCreateFileSuccess(t *testing.T) {
	router := newFilesRouter()

	obj := createTestFile(t, router, "data.jsonl", "batch", []byte(`{"foo":"bar"}`))
	assert.Equal(t, "file", obj.Object)
	assert.Equal(t, "batch", obj.Purpose)
	assert.Equal(t, "data.jsonl", obj.Filename)
	assert.Equal(t, "processed", obj.Status)
	assert.NotEmpty(t, obj.ID)
}

func TestHandleCreateFileMissingFile(t *testing.T) {
	router := newFilesRouter()

	req := newMultipartRequest(t, "/v1/files", map[string]string{"purpose": "batch"}, "", "", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "file parameter")
}

func TestHandleCreateFileMissingPurpose(t *testing.T) {
	router := newFilesRouter()

	req := newMultipartRequest(t, "/v1/files", map[string]string{}, "file", "data.jsonl", []byte("{}"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "purpose parameter")
}

func TestHandleCreateFileInvalidPurpose(t *testing.T) {
	router := newFilesRouter()

	req := newMultipartRequest(t, "/v1/files", map[string]string{"purpose": "not-a-real-purpose"}, "file", "data.jsonl", []byte("{}"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "not a valid purpose")
}

func TestHandleListFilesFilterByPurpose(t *testing.T) {
	router := newFilesRouter()

	batchFile := createTestFile(t, router, "b.jsonl", "batch", []byte("{}"))
	createTestFile(t, router, "v.png", "vision", []byte("fake-png"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/files?purpose=batch", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var body struct {
		Object string              `json:"object"`
		Data   []models.FileObject `json:"data"`
	}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "list", body.Object)
	for _, f := range body.Data {
		assert.Equal(t, "batch", f.Purpose)
	}
	found := false
	for _, f := range body.Data {
		if f.ID == batchFile.ID {
			found = true
		}
	}
	assert.True(t, found)
}

func TestHandleGetFileFound(t *testing.T) {
	router := newFilesRouter()
	obj := createTestFile(t, router, "data.jsonl", "batch", []byte("{}"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/files/"+obj.ID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var fetched models.FileObject
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &fetched))
	assert.Equal(t, obj.ID, fetched.ID)
}

func TestHandleGetFileNotFound(t *testing.T) {
	router := newFilesRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/files/file-does-not-exist", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleDeleteFile(t *testing.T) {
	router := newFilesRouter()
	obj := createTestFile(t, router, "data.jsonl", "batch", []byte("{}"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/v1/files/"+obj.ID, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"deleted":true`)

	// Second delete of the same id should now 404.
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodDelete, "/v1/files/"+obj.ID, nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusNotFound, w2.Code)
}

func TestHandleDeleteFileNotFound(t *testing.T) {
	router := newFilesRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/v1/files/file-does-not-exist", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleDownloadFileContent(t *testing.T) {
	router := newFilesRouter()
	obj := createTestFile(t, router, "data.jsonl", "batch", []byte(`{"foo":"bar"}`))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/files/"+obj.ID+"/content", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/jsonl", w.Header().Get("Content-Type"))
	assert.Equal(t, `{"foo":"bar"}`, w.Body.String())
}

func TestHandleDownloadFileContentNotFound(t *testing.T) {
	router := newFilesRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/files/file-does-not-exist/content", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
