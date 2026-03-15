package mockdata

import (
	"math/rand/v2"

	"mock-openai/internal/models"
)

func GenerateMockEmbedding() []float64 {
	const dimensions = 1536
	vector := make([]float64, dimensions)

	for i := 0; i < dimensions; i++ {
		vector[i] = rand.Float64()*2 - 1
	}

	return vector
}

func GenerateMockEmbeddingResponse(model string, input interface{}) models.EmbeddingResponse {
	var data []models.EmbeddingData
	
	// Type assertion to check if input is a string or a slice
	switch v := input.(type) {
	case string:
		// Single string -> One embedding
		data = append(data, models.EmbeddingData{
			Object:    "embedding",
			Index:     0,
			Embedding: GenerateMockEmbedding(), // Your 1536-dim function
		})
	case []interface{}:
		// JSON array of strings -> Multiple embeddings
		for i := range v {
			data = append(data, models.EmbeddingData{
				Object:    "embedding",
				Index:     i,
				Embedding: GenerateMockEmbedding(),
			})
		}
	}

	return models.EmbeddingResponse{
		Object: "list",
		Data:   data,
		Model:  model,
		Usage: models.UsageStats{
			TotalTokens: 15, // Mock static usage
		},
	}
}