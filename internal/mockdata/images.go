package mockdata

import (
	"mock-openai/internal/models"
	"time"
)

func GenerateMockImageResponse(model string, prompt string) models.ImageResponse {
	return models.ImageResponse{
		Created: time.Now().Unix(),
		Data: []models.ImageData{
			{
				// Using a placeholder service to simulate a generated image
				URL:           "https://picsum.photos/1024",
				RevisedPrompt: "A high-resolution version of: " + prompt,
			},
		},
	}
}
