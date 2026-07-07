package mockdata

import (
	"fmt"
	"time"

	"mock-openai/internal/models"
)

// tinyPNGBase64 is a valid 1x1 transparent PNG, used as a placeholder
// b64_json payload — this mock does not actually render pixels.
const tinyPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII="

const placeholderImageURL = "https://picsum.photos/1024"

// GenerateMockImageResponse builds an ImagesResponse for image
// generation/edits/variations. `n` and `responseFormat` are honored;
// `revisedPrompt`, when non-empty, is only set for dall-e-3 (the only
// model that actually rewrites prompts in the real API).
func GenerateMockImageResponse(model, prompt, responseFormat string, n int) models.ImageResponse {
	if n <= 0 {
		n = 1
	}
	if model == "dall-e-3" && n > 1 {
		// dall-e-3 only supports n=1 in the real API.
		n = 1
	}

	data := make([]models.ImageData, n)
	for i := 0; i < n; i++ {
		var img models.ImageData
		if responseFormat == "b64_json" {
			img.B64JSON = tinyPNGBase64
		} else {
			img.URL = fmt.Sprintf("%s?seed=%d", placeholderImageURL, time.Now().UnixNano()+int64(i))
		}
		if model == "dall-e-3" {
			img.RevisedPrompt = "A high-resolution rendering of: " + prompt
		}
		data[i] = img
	}

	return models.ImageResponse{
		Created: time.Now().Unix(),
		Data:    data,
	}
}
