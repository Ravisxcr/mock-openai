package models

// EmbeddingRequest represents the incoming JSON body for embedding requests.
type EmbeddingRequest struct {
	Input interface{} `json:"input"`
	Model string   `json:"model"`
}

// EmbeddingData represents a single embedding in the response.
type EmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// EmbeddingResponse is the full JSON object returned to the client for embeddings.
type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage   UsageStats      `json:"usage"`
}