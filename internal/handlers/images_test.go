package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"mock-openai/internal/models"
)

func newImagesRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/v1/images/generations", HandleImageGeneration)
	r.POST("/v1/images/edits", HandleImageEdit)
	r.POST("/v1/images/variations", HandleImageVariation)
	return r
}

// newMultipartRequest builds a multipart/form-data request with the given
// fields and an optional file field (fileField/fileName/fileContent).
func newMultipartRequest(t *testing.T, path string, fields map[string]string, fileField, fileName string, fileContent []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		assert.NoError(t, w.WriteField(k, v))
	}
	if fileField != "" {
		fw, err := w.CreateFormFile(fileField, fileName)
		assert.NoError(t, err)
		_, err = fw.Write(fileContent)
		assert.NoError(t, err)
	}
	assert.NoError(t, w.Close())

	req, _ := http.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func TestHandleImageGenerationSuccess(t *testing.T) {
	router := newImagesRouter()

	w := postJSON(t, router, "/v1/images/generations", models.ImageRequest{
		Prompt: "a cat riding a skateboard",
		Model:  "dall-e-3",
		N:      1,
	})
	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.ImageResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	if assert.Len(t, resp.Data, 1) {
		assert.NotEmpty(t, resp.Data[0].URL)
		assert.Contains(t, resp.Data[0].RevisedPrompt, "a cat riding a skateboard")
	}
}

func TestHandleImageGenerationDefaultsModel(t *testing.T) {
	router := newImagesRouter()

	w := postJSON(t, router, "/v1/images/generations", models.ImageRequest{
		Prompt: "a dog",
	})
	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.ImageResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	if assert.Len(t, resp.Data, 1) {
		assert.Empty(t, resp.Data[0].RevisedPrompt, "dall-e-2 (default) should not set revised_prompt")
	}
}

func TestHandleImageGenerationB64JSON(t *testing.T) {
	router := newImagesRouter()

	w := postJSON(t, router, "/v1/images/generations", models.ImageRequest{
		Prompt:         "a sunset",
		ResponseFormat: "b64_json",
	})
	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.ImageResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	if assert.Len(t, resp.Data, 1) {
		assert.NotEmpty(t, resp.Data[0].B64JSON)
		assert.Empty(t, resp.Data[0].URL)
	}
}

func TestHandleImageGenerationMissingPrompt(t *testing.T) {
	router := newImagesRouter()

	w := postJSON(t, router, "/v1/images/generations", models.ImageRequest{Model: "dall-e-2"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "prompt parameter")
}

func TestHandleImageGenerationInvalidJSON(t *testing.T) {
	router := newImagesRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/images/generations", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleImageEditSuccess(t *testing.T) {
	router := newImagesRouter()

	req := newMultipartRequest(t, "/v1/images/edits",
		map[string]string{"prompt": "add a hat"}, "image", "cat.png", []byte("fake-png-bytes"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.ImageResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
}

func TestHandleImageEditMissingPrompt(t *testing.T) {
	router := newImagesRouter()

	req := newMultipartRequest(t, "/v1/images/edits", map[string]string{}, "image", "cat.png", []byte("fake"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "prompt parameter")
}

func TestHandleImageEditMissingImage(t *testing.T) {
	router := newImagesRouter()

	req := newMultipartRequest(t, "/v1/images/edits", map[string]string{"prompt": "add a hat"}, "", "", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "image file")
}

func TestHandleImageVariationSuccess(t *testing.T) {
	router := newImagesRouter()

	req := newMultipartRequest(t, "/v1/images/variations", map[string]string{}, "image", "cat.png", []byte("fake-png-bytes"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.ImageResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
}

func TestHandleImageVariationMissingImage(t *testing.T) {
	router := newImagesRouter()

	req := newMultipartRequest(t, "/v1/images/variations", map[string]string{}, "", "", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "image file")
}
