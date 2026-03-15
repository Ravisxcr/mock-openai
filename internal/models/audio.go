package models

// AudioSpeechRequest is for Text-to-Speech
type AudioSpeechRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
	Voice string `json:"voice"`
}

// AudioResponse is used for Transcriptions and Translations
type AudioResponse struct {
	Text string `json:"text"`
}

// VoiceConsent represents the legal/permission entity
type VoiceConsent struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	CreatedAt int64  `json:"created_at"`
	Status    string `json:"status"`
}
