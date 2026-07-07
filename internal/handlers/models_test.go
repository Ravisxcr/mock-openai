package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newModelsRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/v1/models", HandleModels)
	r.GET("/v1/models/:model_id", HandleModel)
	return r
}

func TestHandleModelsList(t *testing.T) {
	router := newModelsRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/models", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body struct {
		Object string      `json:"object"`
		Data   []modelInfo `json:"data"`
	}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "list", body.Object)
	assert.NotEmpty(t, body.Data)

	found := false
	for _, m := range body.Data {
		if m.ID == "gpt-4o" {
			found = true
			assert.Equal(t, "model", m.Object)
			assert.Equal(t, "system", m.OwnedBy)
		}
	}
	assert.True(t, found, "expected gpt-4o in models list")
}

func TestHandleModelFound(t *testing.T) {
	router := newModelsRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/models/gpt-4o", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var m modelInfo
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &m))
	assert.Equal(t, "gpt-4o", m.ID)
}

func TestHandleModelNotFound(t *testing.T) {
	router := newModelsRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/models/does-not-exist", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errResp struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &errResp))
	assert.Equal(t, "invalid_request_error", errResp.Error.Type)
	assert.Contains(t, errResp.Error.Message, "does-not-exist")
}
