package mockdata

import (
	"math/rand"
	"time"

	"mock-openai/internal/models"
)

func GenerateMockEmbeddingResponse(modelName string, input []string) models.EmbeddingResponse {
	const dimensions = 1536
	vector := make([]float64, dimensions)

	// Seed the random number generator
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < dimensions; i++ {
		// OpenAI embeddings are usually normalized between -1 and 1
		vector[i] = r.Float64()*2 - 1 
	}

	return models.EmbeddingResponse{
		Object: "list",
		Data: []models.EmbeddingData{
			{
				Object:    "embedding",
				Embedding: vector,
				Index:     0,
			},
		},
		Model: modelName,
	}
}