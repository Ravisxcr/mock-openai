// Package store holds simple in-memory state for endpoints that let
// clients create then later retrieve/list/delete a resource (e.g. the
// stored chat completions endpoints). It is not persisted across restarts.
package store

import (
	"sync"

	"mock-openai/internal/models"
)

var (
	mu          sync.Mutex
	completions = map[string]*models.ChatResponse{}
	order       []string
)

// SaveCompletion stores (or overwrites) a completion by id.
func SaveCompletion(resp models.ChatResponse) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := completions[resp.ID]; !exists {
		order = append(order, resp.ID)
	}
	completions[resp.ID] = &resp
}

// GetCompletion returns the stored completion, if any.
func GetCompletion(id string) (models.ChatResponse, bool) {
	mu.Lock()
	defer mu.Unlock()
	resp, ok := completions[id]
	if !ok {
		return models.ChatResponse{}, false
	}
	return *resp, true
}

// DeleteCompletion removes a stored completion, reporting whether it existed.
func DeleteCompletion(id string) bool {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := completions[id]; !ok {
		return false
	}
	delete(completions, id)
	for i, existing := range order {
		if existing == id {
			order = append(order[:i], order[i+1:]...)
			break
		}
	}
	return true
}

// ListCompletions returns stored completions, most recently created first.
func ListCompletions() []models.ChatResponse {
	mu.Lock()
	defer mu.Unlock()
	result := make([]models.ChatResponse, 0, len(order))
	for i := len(order) - 1; i >= 0; i-- {
		result = append(result, *completions[order[i]])
	}
	return result
}
