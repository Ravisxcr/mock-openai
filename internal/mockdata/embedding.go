package mockdata

import (
	"encoding/base64"
	"encoding/binary"
	"math"
	"math/rand/v2"
	"strings"

	"mock-openai/internal/models"
)

// defaultDimensions mirrors the real per-model embedding size.
func defaultDimensions(model string) int {
	switch model {
	case "text-embedding-3-large":
		return 3072
	default: // text-embedding-3-small, text-embedding-ada-002, and anything unknown
		return 1536
	}
}

func generateVector(dimensions int) []float32 {
	vector := make([]float32, dimensions)
	for i := range vector {
		vector[i] = rand.Float32()*2 - 1
	}
	return vector
}

// encodeEmbedding renders a vector as either a []float32 or, for
// encoding_format=base64, a base64 string of its little-endian float32 bytes
// (matching real API behavior).
func encodeEmbedding(vector []float32, base64Encoding bool) interface{} {
	if !base64Encoding {
		return vector
	}
	buf := make([]byte, 4*len(vector))
	for i, v := range vector {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
	}
	return base64.StdEncoding.EncodeToString(buf)
}

func estimateEmbeddingTokens(input interface{}) int {
	switch v := input.(type) {
	case string:
		return estimateTokens(v)
	case []interface{}:
		var all []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				all = append(all, s)
			}
		}
		return estimateTokens(strings.Join(all, " "))
	default:
		return 0
	}
}

func GenerateMockEmbeddingResponse(req models.EmbeddingRequest) models.EmbeddingResponse {
	dimensions := defaultDimensions(req.Model)
	if req.Dimensions != nil && *req.Dimensions > 0 {
		dimensions = *req.Dimensions
	}
	base64Encoding := req.EncodingFormat == "base64"

	var data []models.EmbeddingData
	switch v := req.Input.(type) {
	case string:
		data = append(data, models.EmbeddingData{
			Object:    "embedding",
			Index:     0,
			Embedding: encodeEmbedding(generateVector(dimensions), base64Encoding),
		})
	case []interface{}:
		for i := range v {
			data = append(data, models.EmbeddingData{
				Object:    "embedding",
				Index:     i,
				Embedding: encodeEmbedding(generateVector(dimensions), base64Encoding),
			})
		}
	}

	promptTokens := estimateEmbeddingTokens(req.Input)
	return models.EmbeddingResponse{
		Object: "list",
		Data:   data,
		Model:  req.Model,
		Usage: models.EmbeddingUsage{
			PromptTokens: promptTokens,
			TotalTokens:  promptTokens,
		},
	}
}
