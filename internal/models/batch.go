package models

// BatchRequestCounts mirrors the `request_counts` object on `Batch`.
type BatchRequestCounts struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}

// CreateBatchRequest is the body sent by the client to POST /v1/batches.
type CreateBatchRequest struct {
	InputFileID      string            `json:"input_file_id"`
	Endpoint         string            `json:"endpoint"`
	CompletionWindow string            `json:"completion_window"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// Batch mirrors the OpenAI `Batch` schema.
type Batch struct {
	ID               string             `json:"id"`
	Object           string             `json:"object"` // always "batch"
	Endpoint         string             `json:"endpoint"`
	Errors           interface{}        `json:"errors"`
	InputFileID      string             `json:"input_file_id"`
	CompletionWindow string             `json:"completion_window"`
	Status           string             `json:"status"`
	OutputFileID     *string            `json:"output_file_id,omitempty"`
	ErrorFileID      *string            `json:"error_file_id,omitempty"`
	CreatedAt        int64              `json:"created_at"`
	InProgressAt     *int64             `json:"in_progress_at,omitempty"`
	ExpiresAt        *int64             `json:"expires_at,omitempty"`
	FinalizingAt     *int64             `json:"finalizing_at,omitempty"`
	CompletedAt      *int64             `json:"completed_at,omitempty"`
	FailedAt         *int64             `json:"failed_at,omitempty"`
	ExpiredAt        *int64             `json:"expired_at,omitempty"`
	CancellingAt     *int64             `json:"cancelling_at,omitempty"`
	CancelledAt      *int64             `json:"cancelled_at,omitempty"`
	RequestCounts    BatchRequestCounts `json:"request_counts"`
	Metadata         map[string]string  `json:"metadata,omitempty"`
}

// ListBatchesResponse mirrors `ListBatchesResponse`.
type ListBatchesResponse struct {
	Object  string  `json:"object"`
	Data    []Batch `json:"data"`
	FirstID string  `json:"first_id,omitempty"`
	LastID  string  `json:"last_id,omitempty"`
	HasMore bool    `json:"has_more"`
}
