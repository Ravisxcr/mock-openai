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

// HandleAudioSpeech generates audio from the input text.
//
// @Summary		Create speech
// @Description	Generates audio from the input text. Returns a fixed mock audio payload with a content-type matching the requested response_format.
// @Tags			Audio
// @Accept			json
// @Produce		audio/mpeg
// @Param			request	body		models.AudioSpeechRequest	true	"Speech request"
// @Success		200		{file}		binary
// @Failure		400		{object}	errs.Response
// @Failure		401	{object}	errs.Response
// @Security		BearerAuth
// @Router			/audio/speech [post]
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

// HandleAudioTranscription transcribes audio into the input language.
//
// @Summary		Create transcription
// @Description	Transcribes audio into the input language. response_format controls the reply shape: json (default), text, srt, vtt, or verbose_json.
// @Tags			Audio
// @Accept			multipart/form-data
// @Produce		json
// @Param			file			formData	file	true	"Audio file to transcribe"
// @Param			model			formData	string	true	"Model ID (e.g. whisper-1)"
// @Param			response_format	formData	string	false	"json, text, srt, vtt, or verbose_json"
// @Success		200				{object}	models.AudioResponse
// @Failure		400				{object}	errs.Response
// @Failure		401				{object}	errs.Response
// @Security		BearerAuth
// @Router			/audio/transcriptions [post]
func HandleAudioTranscription(c *gin.Context) {
	handleAudioInput(c, "This is a mock transcription of your audio file.")
}

// HandleAudioTranslation translates audio into English.
//
// @Summary		Create translation
// @Description	Translates audio into English. response_format controls the reply shape: json (default), text, srt, vtt, or verbose_json.
// @Tags			Audio
// @Accept			multipart/form-data
// @Produce		json
// @Param			file			formData	file	true	"Audio file to translate"
// @Param			model			formData	string	true	"Model ID (e.g. whisper-1)"
// @Param			response_format	formData	string	false	"json, text, srt, vtt, or verbose_json"
// @Success		200				{object}	models.AudioResponse
// @Failure		400				{object}	errs.Response
// @Failure		401				{object}	errs.Response
// @Security		BearerAuth
// @Router			/audio/translations [post]
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

// HandleListVoiceConsents lists voice consents.
//
// Voice Consent Handlers (approximate — not verified against VoiceConsentResource)
//
// @Summary		List voice consents
// @Tags			Audio
// @Produce		json
// @Success		200	{object}	object{object=string,data=[]models.VoiceConsent}
// @Failure		401	{object}	errs.Response
// @Security		BearerAuth
// @Router			/audio/voice_consents [get]
func HandleListVoiceConsents(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"object": "list", "data": []models.VoiceConsent{}})
}

// HandleCreateVoiceConsent creates (or updates) a voice consent. It also
// serves as the update handler for POST /audio/voice_consents/{consent_id}.
//
// @Summary		Create voice consent
// @Tags			Audio
// @Produce		json
// @Param			consent_id	path		string	false	"Voice consent ID (update variant only)"
// @Success		201			{object}	models.VoiceConsent
// @Failure		401			{object}	errs.Response
// @Security		BearerAuth
// @Router			/audio/voice_consents [post]
// @Router			/audio/voice_consents/{consent_id} [post]
func HandleCreateVoiceConsent(c *gin.Context) {
	c.JSON(http.StatusCreated, models.VoiceConsent{ID: "vc_123", Object: "voice_consent", CreatedAt: time.Now().Unix(), Status: "active"})
}

// HandleGetVoiceConsent retrieves a voice consent by id.
//
// @Summary		Get voice consent
// @Tags			Audio
// @Produce		json
// @Param			consent_id	path		string	true	"Voice consent ID"
// @Success		200			{object}	models.VoiceConsent
// @Failure		401			{object}	errs.Response
// @Security		BearerAuth
// @Router			/audio/voice_consents/{consent_id} [get]
func HandleGetVoiceConsent(c *gin.Context) {
	id := c.Param("consent_id")
	c.JSON(http.StatusOK, models.VoiceConsent{ID: id, Object: "voice_consent", Status: "active"})
}

// HandleDeleteVoiceConsent deletes a voice consent by id.
//
// @Summary		Delete voice consent
// @Tags			Audio
// @Produce		json
// @Param			consent_id	path		string	true	"Voice consent ID"
// @Success		200			{object}	object{id=string,object=string,deleted=bool}
// @Failure		401			{object}	errs.Response
// @Security		BearerAuth
// @Router			/audio/voice_consents/{consent_id} [delete]
func HandleDeleteVoiceConsent(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"id": c.Param("consent_id"), "object": "voice_consent", "deleted": true})
}
