package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"mock-openai/internal/apierror"
	"mock-openai/internal/mockdata"
	"mock-openai/internal/store"
)

var validFilePurposes = map[string]bool{
	"assistants": true,
	"batch":      true,
	"fine-tune":  true,
	"vision":     true,
	"user_data":  true,
	"evals":      true,
}

// POST /v1/files (multipart/form-data: file, purpose, ...)
func HandleCreateFile(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		apierror.InvalidRequest(c, "you must provide a file parameter", "file")
		return
	}
	purpose := c.PostForm("purpose")
	if purpose == "" {
		apierror.InvalidRequest(c, "you must provide a purpose parameter", "purpose")
		return
	}
	if !validFilePurposes[purpose] {
		apierror.InvalidRequest(c, fmt.Sprintf("'%s' is not a valid purpose", purpose), "purpose")
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "server_error", "Failed to read uploaded file", "", "")
		return
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "server_error", "Failed to read uploaded file", "", "")
		return
	}

	obj := mockdata.BuildFileObject(fileHeader.Filename, purpose, int64(len(content)))
	store.SaveFile(obj, content)
	c.JSON(http.StatusOK, obj)
}

// GET /v1/files
func HandleListFiles(c *gin.Context) {
	purpose := c.Query("purpose")
	files := store.ListFiles(purpose)
	firstID, lastID := "", ""
	if len(files) > 0 {
		firstID, lastID = files[0].ID, files[len(files)-1].ID
	}
	c.JSON(http.StatusOK, gin.H{
		"object":   "list",
		"data":     files,
		"first_id": firstID,
		"last_id":  lastID,
		"has_more": false,
	})
}

// GET /v1/files/:file_id
func HandleGetFile(c *gin.Context) {
	id := c.Param("file_id")
	obj, ok := store.GetFile(id)
	if !ok {
		apierror.NotFound(c, fmt.Sprintf("No such file: '%s'", id), "file_not_found")
		return
	}
	c.JSON(http.StatusOK, obj)
}

// DELETE /v1/files/:file_id
func HandleDeleteFile(c *gin.Context) {
	id := c.Param("file_id")
	if !store.DeleteFile(id) {
		apierror.NotFound(c, fmt.Sprintf("No such file: '%s'", id), "file_not_found")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"object":  "file",
		"deleted": true,
	})
}

// GET /v1/files/:file_id/content
func HandleDownloadFileContent(c *gin.Context) {
	id := c.Param("file_id")
	obj, ok := store.GetFile(id)
	if !ok {
		apierror.NotFound(c, fmt.Sprintf("No such file: '%s'", id), "file_not_found")
		return
	}
	content, _ := store.GetFileContent(id)

	contentType := "application/octet-stream"
	if strings.HasSuffix(obj.Filename, ".jsonl") || obj.Purpose == "batch" || obj.Purpose == "batch_output" {
		contentType = "application/jsonl"
	}
	c.Data(http.StatusOK, contentType, content)
}
