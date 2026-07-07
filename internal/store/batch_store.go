package store

import (
	"sync"

	"mock-openai/internal/models"
)

var (
	batchMu    sync.Mutex
	batches    = map[string]*models.Batch{}
	batchOrder []string
)

// SaveBatch stores (or overwrites) a batch by id.
func SaveBatch(b models.Batch) {
	batchMu.Lock()
	defer batchMu.Unlock()
	if _, exists := batches[b.ID]; !exists {
		batchOrder = append(batchOrder, b.ID)
	}
	batches[b.ID] = &b
}

// GetBatch returns the stored batch, if any.
func GetBatch(id string) (models.Batch, bool) {
	batchMu.Lock()
	defer batchMu.Unlock()
	b, ok := batches[id]
	if !ok {
		return models.Batch{}, false
	}
	return *b, true
}

// ListBatches returns stored batches, most recently created first.
func ListBatches() []models.Batch {
	batchMu.Lock()
	defer batchMu.Unlock()
	result := make([]models.Batch, 0, len(batchOrder))
	for i := len(batchOrder) - 1; i >= 0; i-- {
		result = append(result, *batches[batchOrder[i]])
	}
	return result
}
