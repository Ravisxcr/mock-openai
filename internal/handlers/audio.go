package handlers

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"mock-openai/internal/errs"
	"mock-openai/internal/models"
)

const mockAudioBase64 = `SUQzBAAAAAAAF1RTU0UAAAANAAADTGFtZTMuMTAwVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVf/4xAAYAAAaA8AAAACAAAnS`

var speechContentTypes = map[string]string{
	"mp3":  "audio/mpeg",
	"opus": "audio/opus",
	"aac":  "audio/aac",
	"flac": "audio/flac",
	"wav":  "audio/wav",
	"pcm":  "audio/pcm",
}

// POST /v1/audio/speech
func HandleAudioSpeech(c *gin.Context) {
	var req models.AudioSpeechRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errs.InvalidRequest(c, "Invalid request body: "+err.Error(), "")
		return
	}
	if req.Model == "" {
		errs.InvalidRequest(c, "you must provide a model parameter", "model")
		return
	}
	if req.Input == "" {
		errs.InvalidRequest(c, "you must provide an input parameter", "input")
		return
	}
	if req.Voice == "" {
		errs.InvalidRequest(c, "you must provide a voice parameter", "voice")
		return
	}

	contentType, ok := speechContentTypes[req.ResponseFormat]
	if !ok {
		contentType = "audio/mpeg" // default (response_format defaults to mp3)
	}

	audioBytes, err := base64.StdEncoding.DecodeString(mockAudioBase64)
	if err != nil {
		errs.Respond(c, http.StatusInternalServerError, "server_error", "Failed to generate mock audio", "", "")
		return
	}
	c.Data(http.StatusOK, contentType, audioBytes)
}

// POST /v1/audio/transcriptions (multipart/form-data: file, model, ...)
func HandleAudioTranscription(c *gin.Context) {
	handleAudioInput(c, "This is a mock transcription of your audio file.")
}

// POST /v1/audio/translations (multipart/form-data: file, model, ...)
func HandleAudioTranslation(c *gin.Context) {
	handleAudioInput(c, "This is a mock translation into English.")
}

func handleAudioInput(c *gin.Context, text string) {
	if _, _, err := c.Request.FormFile("file"); err != nil {
		errs.InvalidRequest(c, "you must provide a file parameter", "file")
		return
	}
	if c.PostForm("model") == "" {
		errs.InvalidRequest(c, "you must provide a model parameter", "model")
		return
	}

	switch c.PostForm("response_format") {
	case "text":
		c.String(http.StatusOK, text)
	case "srt":
		c.String(http.StatusOK, "1\n00:00:00,000 --> 00:00:03,000\n%s\n", text)
	case "vtt":
		c.String(http.StatusOK, "WEBVTT\n\n00:00:00.000 --> 00:00:03.000\n%s\n", text)
	case "verbose_json":
		c.JSON(http.StatusOK, models.AudioVerboseResponse{
			Language: "english",
			Duration: 3.0,
			Text:     text,
			Segments: []interface{}{},
			Words:    []interface{}{},
		})
	default:
		c.JSON(http.StatusOK, models.AudioResponse{Text: text})
	}
}

// Voice Consent Handlers (approximate — not verified against VoiceConsentResource)
func HandleListVoiceConsents(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"object": "list", "data": []models.VoiceConsent{}})
}

func HandleCreateVoiceConsent(c *gin.Context) {
	c.JSON(http.StatusCreated, models.VoiceConsent{ID: "vc_123", Object: "voice_consent", CreatedAt: time.Now().Unix(), Status: "active"})
}

func HandleGetVoiceConsent(c *gin.Context) {
	id := c.Param("consent_id")
	c.JSON(http.StatusOK, models.VoiceConsent{ID: id, Object: "voice_consent", Status: "active"})
}

func HandleDeleteVoiceConsent(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"id": c.Param("consent_id"), "object": "voice_consent", "deleted": true})
}
