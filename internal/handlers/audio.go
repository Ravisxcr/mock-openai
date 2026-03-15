package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"mock-openai/internal/models"
)

// POST /v1/audio/transcriptions
func HandleAudioTranscription(c *gin.Context) {
	c.JSON(http.StatusOK, models.AudioResponse{Text: "This is a mock transcription of your audio file."})
}

// POST /v1/audio/translations
func HandleAudioTranslation(c *gin.Context) {
	c.JSON(http.StatusOK, models.AudioResponse{Text: "This is a mock translation into English."})
}

// POST /v1/audio/speech
func HandleAudioSpeech(c *gin.Context) {
	// OpenAI returns binary audio data (MP3/WAV)
	c.Header("Content-Type", "audio/mpeg")
	// Return a tiny 1-second silence or empty byte slice for mocking
	c.Data(http.StatusOK, "audio/mpeg", []byte{0x00}) 
}

// Voice Consent Handlers
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