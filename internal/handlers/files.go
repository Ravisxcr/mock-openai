package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"mock-openai/internal/errs"
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

// HandleCreateFile uploads a file for use with batch, fine-tuning, assistants, etc.
//
// @Summary		Upload file
// @Description	Uploads a file that can be used across various endpoints (assistants, batch, fine-tune, vision, user_data, evals).
// @Tags			Files
// @Accept			multipart/form-data
// @Produce		json
// @Param			file	formData	file	true	"File to upload"
// @Param			purpose	formData	string	true	"assistants, batch, fine-tune, vision, user_data, or evals"
// @Success		200		{object}	models.FileObject
// @Failure		400		{object}	errs.Response
// @Failure		401		{object}	errs.Response
// @Security		BearerAuth
// @Router			/files [post]
func HandleCreateFile(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		errs.InvalidRequest(c, "you must provide a file parameter", "file")
		return
	}
	purpose := c.PostForm("purpose")
	if purpose == "" {
		errs.InvalidRequest(c, "you must provide a purpose parameter", "purpose")
		return
	}
	if !validFilePurposes[purpose] {
		errs.InvalidRequest(c, fmt.Sprintf("'%s' is not a valid purpose", purpose), "purpose")
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		errs.Respond(c, http.StatusInternalServerError, "server_error", "Failed to read uploaded file", "", "")
		return
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		errs.Respond(c, http.StatusInternalServerError, "server_error", "Failed to read uploaded file", "", "")
		return
	}

	obj := mockdata.BuildFileObject(fileHeader.Filename, purpose, int64(len(content)))
	store.SaveFile(obj, content)
	c.JSON(http.StatusOK, obj)
}

// HandleListFiles lists uploaded files.
//
// @Summary		List files
// @Description	Returns a list of files, optionally filtered by purpose.
// @Tags			Files
// @Produce		json
// @Param			purpose	query		string	false	"Only return files with this purpose"
// @Success		200		{object}	models.ListFilesResponse
// @Failure		401		{object}	errs.Response
// @Security		BearerAuth
// @Router			/files [get]
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

// HandleGetFile retrieves a file's metadata by id.
//
// @Summary		Get file
// @Description	Returns information about a specific file.
// @Tags			Files
// @Produce		json
// @Param			file_id	path		string	true	"File ID"
// @Success		200		{object}	models.FileObject
// @Failure		404		{object}	errs.Response
// @Security		BearerAuth
// @Router			/files/{file_id} [get]
func HandleGetFile(c *gin.Context) {
	id := c.Param("file_id")
	obj, ok := store.GetFile(id)
	if !ok {
		errs.NotFound(c, fmt.Sprintf("No such file: '%s'", id), "file_not_found")
		return
	}
	c.JSON(http.StatusOK, obj)
}

// HandleDeleteFile deletes a file by id.
//
// @Summary		Delete file
// @Tags			Files
// @Produce		json
// @Param			file_id	path		string	true	"File ID"
// @Success		200		{object}	models.DeleteFileResponse
// @Failure		404		{object}	errs.Response
// @Security		BearerAuth
// @Router			/files/{file_id} [delete]
func HandleDeleteFile(c *gin.Context) {
	id := c.Param("file_id")
	if !store.DeleteFile(id) {
		errs.NotFound(c, fmt.Sprintf("No such file: '%s'", id), "file_not_found")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"object":  "file",
		"deleted": true,
	})
}

// HandleDownloadFileContent downloads the raw content of a file.
//
// @Summary		Download file content
// @Tags			Files
// @Produce		application/octet-stream
// @Param			file_id	path	string	true	"File ID"
// @Success		200		{file}	binary
// @Failure		404		{object}	errs.Response
// @Security		BearerAuth
// @Router			/files/{file_id}/content [get]
func HandleDownloadFileContent(c *gin.Context) {
	id := c.Param("file_id")
	obj, ok := store.GetFile(id)
	if !ok {
		errs.NotFound(c, fmt.Sprintf("No such file: '%s'", id), "file_not_found")
		return
	}
	content, _ := store.GetFileContent(id)

	contentType := "application/octet-stream"
	if strings.HasSuffix(obj.Filename, ".jsonl") || obj.Purpose == "batch" || obj.Purpose == "batch_output" {
		contentType = "application/jsonl"
	}
	c.Data(http.StatusOK, contentType, content)
}
