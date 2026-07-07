package mockdata

import (
	"time"

	"mock-openai/internal/models"
)

// BuildFileObject constructs a FileObject for a newly uploaded (or
// server-generated, e.g. batch output) file.
func BuildFileObject(filename, purpose string, size int64) models.FileObject {
	return models.FileObject{
		ID:        "file-" + randomSuffix(),
		Object:    "file",
		Bytes:     size,
		CreatedAt: time.Now().Unix(),
		Filename:  filename,
		Purpose:   purpose,
		Status:    "processed",
	}
}
