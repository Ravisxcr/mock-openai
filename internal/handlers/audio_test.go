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

func newAudioRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/v1/audio/speech", HandleAudioSpeech)
	r.POST("/v1/audio/transcriptions", HandleAudioTranscription)
	r.POST("/v1/audio/translations", HandleAudioTranslation)
	r.GET("/v1/audio/voice_consents", HandleListVoiceConsents)
	r.POST("/v1/audio/voice_consents", HandleCreateVoiceConsent)
	r.GET("/v1/audio/voice_consents/:consent_id", HandleGetVoiceConsent)
	r.POST("/v1/audio/voice_consents/:consent_id", HandleCreateVoiceConsent)
	r.DELETE("/v1/audio/voice_consents/:consent_id", HandleDeleteVoiceConsent)
	return r
}

func TestHandleAudioSpeechSuccessDefaultFormat(t *testing.T) {
	router := newAudioRouter()

	w := postJSON(t, router, "/v1/audio/speech", models.AudioSpeechRequest{
		Model: "tts-1",
		Input: "Hello there",
		Voice: "alloy",
	})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "audio/mpeg", w.Header().Get("Content-Type"))
	assert.NotEmpty(t, w.Body.Bytes())
}

func TestHandleAudioSpeechCustomFormat(t *testing.T) {
	router := newAudioRouter()

	w := postJSON(t, router, "/v1/audio/speech", models.AudioSpeechRequest{
		Model:          "tts-1",
		Input:          "Hello there",
		Voice:          "alloy",
		ResponseFormat: "wav",
	})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "audio/wav", w.Header().Get("Content-Type"))
}

func TestHandleAudioSpeechValidation(t *testing.T) {
	router := newAudioRouter()

	cases := []struct {
		name string
		req  models.AudioSpeechRequest
		want string
	}{
		{"missing model", models.AudioSpeechRequest{Input: "hi", Voice: "alloy"}, "model parameter"},
		{"missing input", models.AudioSpeechRequest{Model: "tts-1", Voice: "alloy"}, "input parameter"},
		{"missing voice", models.AudioSpeechRequest{Model: "tts-1", Input: "hi"}, "voice parameter"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := postJSON(t, router, "/v1/audio/speech", tc.req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), tc.want)
		})
	}
}

func TestHandleAudioTranscriptionDefaultJSON(t *testing.T) {
	router := newAudioRouter()

	req := newMultipartRequest(t, "/v1/audio/transcriptions",
		map[string]string{"model": "whisper-1"}, "file", "audio.mp3", []byte("fake-audio-bytes"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.AudioResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "This is a mock transcription of your audio file.", resp.Text)
}

func TestHandleAudioTranscriptionTextFormat(t *testing.T) {
	router := newAudioRouter()

	req := newMultipartRequest(t, "/v1/audio/transcriptions",
		map[string]string{"model": "whisper-1", "response_format": "text"}, "file", "audio.mp3", []byte("fake"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "This is a mock transcription of your audio file.", w.Body.String())
}

func TestHandleAudioTranscriptionVerboseJSON(t *testing.T) {
	router := newAudioRouter()

	req := newMultipartRequest(t, "/v1/audio/transcriptions",
		map[string]string{"model": "whisper-1", "response_format": "verbose_json"}, "file", "audio.mp3", []byte("fake"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.AudioVerboseResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "english", resp.Language)
}

func TestHandleAudioTranscriptionSRTAndVTT(t *testing.T) {
	router := newAudioRouter()

	for _, format := range []string{"srt", "vtt"} {
		t.Run(format, func(t *testing.T) {
			req := newMultipartRequest(t, "/v1/audio/transcriptions",
				map[string]string{"model": "whisper-1", "response_format": format}, "file", "audio.mp3", []byte("fake"))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.NotEmpty(t, w.Body.String())
		})
	}
}

func TestHandleAudioTranscriptionMissingFile(t *testing.T) {
	router := newAudioRouter()

	req := newMultipartRequest(t, "/v1/audio/transcriptions", map[string]string{"model": "whisper-1"}, "", "", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "file parameter")
}

func TestHandleAudioTranscriptionMissingModel(t *testing.T) {
	router := newAudioRouter()

	req := newMultipartRequest(t, "/v1/audio/transcriptions", map[string]string{}, "file", "audio.mp3", []byte("fake"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "model parameter")
}

func TestHandleAudioTranslationSuccess(t *testing.T) {
	router := newAudioRouter()

	req := newMultipartRequest(t, "/v1/audio/translations",
		map[string]string{"model": "whisper-1"}, "file", "audio.mp3", []byte("fake-audio-bytes"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.AudioResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "This is a mock translation into English.", resp.Text)
}

func TestVoiceConsentLifecycle(t *testing.T) {
	router := newAudioRouter()

	// List (initially returns an empty stub list)
	wList := httptest.NewRecorder()
	reqList, _ := http.NewRequest(http.MethodGet, "/v1/audio/voice_consents", nil)
	router.ServeHTTP(wList, reqList)
	assert.Equal(t, http.StatusOK, wList.Code)
	assert.Contains(t, wList.Body.String(), `"object":"list"`)

	// Create
	wCreate := httptest.NewRecorder()
	reqCreate, _ := http.NewRequest(http.MethodPost, "/v1/audio/voice_consents", nil)
	router.ServeHTTP(wCreate, reqCreate)
	assert.Equal(t, http.StatusCreated, wCreate.Code)
	var created models.VoiceConsent
	assert.NoError(t, json.Unmarshal(wCreate.Body.Bytes(), &created))
	assert.Equal(t, "active", created.Status)
	assert.NotEmpty(t, created.ID)

	// Get by id
	wGet := httptest.NewRecorder()
	reqGet, _ := http.NewRequest(http.MethodGet, "/v1/audio/voice_consents/"+created.ID, nil)
	router.ServeHTTP(wGet, reqGet)
	assert.Equal(t, http.StatusOK, wGet.Code)
	var fetched models.VoiceConsent
	assert.NoError(t, json.Unmarshal(wGet.Body.Bytes(), &fetched))
	assert.Equal(t, created.ID, fetched.ID)

	// Update (POST to :consent_id)
	wUpdate := httptest.NewRecorder()
	reqUpdate, _ := http.NewRequest(http.MethodPost, "/v1/audio/voice_consents/"+created.ID, nil)
	router.ServeHTTP(wUpdate, reqUpdate)
	assert.Equal(t, http.StatusCreated, wUpdate.Code)

	// Delete
	wDelete := httptest.NewRecorder()
	reqDelete, _ := http.NewRequest(http.MethodDelete, "/v1/audio/voice_consents/"+created.ID, nil)
	router.ServeHTTP(wDelete, reqDelete)
	assert.Equal(t, http.StatusOK, wDelete.Code)
	assert.Contains(t, wDelete.Body.String(), `"deleted":true`)
}
