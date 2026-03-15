package handlers

import (
	"net/http"
	"time"
	"encoding/base64"

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
	// In a real project, you could load this from a .mp3 file in your assets folder
	const mockAudioBase64 = `SUQzBAAAAAAAF1RTU0UAAAANAAADTGFtZTMuMTAwVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVf/4xAAYAAAaA8AAAACAAAnS`

	audioBytes, err := base64.StdEncoding.DecodeString(mockAudioBase64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate mock audio"})
		return
	}

	// Set the correct header so the client knows it is receiving an MP3
	c.Header("Content-Type", "audio/mpeg")
	c.Header("Content-Length", string(len(audioBytes)))
	
	// Send the binary data
	c.Data(http.StatusOK, "audio/mpeg", audioBytes)
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