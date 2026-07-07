package models

// FileObject mirrors the OpenAI `OpenAIFile` schema.
type FileObject struct {
	ID        string `json:"id"`
	Object    string `json:"object"` // always "file"
	Bytes     int64  `json:"bytes"`
	CreatedAt int64  `json:"created_at"`
	ExpiresAt *int64 `json:"expires_at,omitempty"`
	Filename  string `json:"filename"`
	Purpose   string `json:"purpose"`
	Status    string `json:"status"` // deprecated in the real API, but required; always "processed" here
}

// ListFilesResponse mirrors `ListFilesResponse`.
type ListFilesResponse struct {
	Object  string       `json:"object"`
	Data    []FileObject `json:"data"`
	FirstID string       `json:"first_id"`
	LastID  string       `json:"last_id"`
	HasMore bool         `json:"has_more"`
}

// DeleteFileResponse mirrors `DeleteFileResponse`.
type DeleteFileResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"` // always "file"
	Deleted bool   `json:"deleted"`
}
