package models

// AudioSpeechRequest is for Text-to-Speech (CreateSpeechRequest schema).
type AudioSpeechRequest struct {
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format"`
	Speed          float64 `json:"speed"`
}

// AudioResponse is used for Transcriptions/Translations JSON responses.
type AudioResponse struct {
	Text string `json:"text"`
}

// AudioVerboseResponse is used when response_format=verbose_json.
type AudioVerboseResponse struct {
	Language string        `json:"language"`
	Duration float64       `json:"duration"`
	Text     string        `json:"text"`
	Segments []interface{} `json:"segments"`
	Words    []interface{} `json:"words"`
}

// VoiceConsent represents the legal/permission entity
type VoiceConsent struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	CreatedAt int64  `json:"created_at"`
	Status    string `json:"status"`
}
