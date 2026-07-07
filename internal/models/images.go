package models

// ImageRequest covers dall-e-2/dall-e-3 prompt-based generation.
type ImageRequest struct {
	Prompt         string `json:"prompt"`
	Model          string `json:"model"`
	N              int    `json:"n"`
	Quality        string `json:"quality"`
	Size           string `json:"size"`
	Style          string `json:"style"`
	ResponseFormat string `json:"response_format"`
	User           string `json:"user"`
}

// ImageResponse is the standard wrapper for image results (ImagesResponse schema).
type ImageResponse struct {
	Created int64       `json:"created"`
	Data    []ImageData `json:"data"`
}

type ImageData struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}
