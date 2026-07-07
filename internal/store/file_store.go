package store

import (
	"sync"

	"mock-openai/internal/models"
)

type storedFile struct {
	obj     models.FileObject
	content []byte
}

var (
	fileMu    sync.Mutex
	files     = map[string]*storedFile{}
	fileOrder []string
)

// SaveFile stores (or overwrites) a file object and its raw content by id.
func SaveFile(obj models.FileObject, content []byte) {
	fileMu.Lock()
	defer fileMu.Unlock()
	if _, exists := files[obj.ID]; !exists {
		fileOrder = append(fileOrder, obj.ID)
	}
	files[obj.ID] = &storedFile{obj: obj, content: content}
}

// GetFile returns the stored file object, if any.
func GetFile(id string) (models.FileObject, bool) {
	fileMu.Lock()
	defer fileMu.Unlock()
	f, ok := files[id]
	if !ok {
		return models.FileObject{}, false
	}
	return f.obj, true
}

// GetFileContent returns the stored raw bytes for a file, if any.
func GetFileContent(id string) ([]byte, bool) {
	fileMu.Lock()
	defer fileMu.Unlock()
	f, ok := files[id]
	if !ok {
		return nil, false
	}
	return f.content, true
}

// DeleteFile removes a stored file, reporting whether it existed.
func DeleteFile(id string) bool {
	fileMu.Lock()
	defer fileMu.Unlock()
	if _, ok := files[id]; !ok {
		return false
	}
	delete(files, id)
	for i, existing := range fileOrder {
		if existing == id {
			fileOrder = append(fileOrder[:i], fileOrder[i+1:]...)
			break
		}
	}
	return true
}

// ListFiles returns stored files, most recently created first, optionally
// filtered by purpose (empty string means no filter).
func ListFiles(purpose string) []models.FileObject {
	fileMu.Lock()
	defer fileMu.Unlock()
	result := make([]models.FileObject, 0, len(fileOrder))
	for i := len(fileOrder) - 1; i >= 0; i-- {
		f := files[fileOrder[i]].obj
		if purpose != "" && f.Purpose != purpose {
			continue
		}
		result = append(result, f)
	}
	return result
}
