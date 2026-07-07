package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"mock-openai/internal/errs"
	"mock-openai/internal/mockdata"
	"mock-openai/internal/models"
	"mock-openai/internal/store"
)

// HandleCreateBatch creates and runs a batch from an uploaded file of requests.
//
// @Summary		Create batch
// @Description	Creates and runs a batch job against a previously uploaded input file. Unlike the real API, batches in this mock complete synchronously and are ready by the time the create call returns.
// @Tags			Batch
// @Accept			json
// @Produce		json
// @Param			request	body		models.CreateBatchRequest	true	"Batch creation request"
// @Success		200		{object}	models.Batch
// @Failure		400		{object}	errs.Response
// @Failure		401		{object}	errs.Response
// @Security		BearerAuth
// @Router			/batches [post]
func HandleCreateBatch(c *gin.Context) {
	var req models.CreateBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errs.InvalidRequest(c, "Invalid request body: "+err.Error(), "")
		return
	}
	if req.InputFileID == "" {
		errs.InvalidRequest(c, "you must provide an input_file_id parameter", "input_file_id")
		return
	}
	if req.Endpoint == "" {
		errs.InvalidRequest(c, "you must provide an endpoint parameter", "endpoint")
		return
	}
	if req.CompletionWindow == "" {
		errs.InvalidRequest(c, "you must provide a completion_window parameter", "completion_window")
		return
	}

	inputContent, ok := store.GetFileContent(req.InputFileID)
	if !ok {
		errs.InvalidRequest(c, fmt.Sprintf("No such file: '%s'", req.InputFileID), "input_file_id")
		return
	}

	batch, outputContent, outputFilename := mockdata.BuildBatch(req, inputContent)

	outputFile := mockdata.BuildFileObject(outputFilename, "batch_output", int64(len(outputContent)))
	store.SaveFile(outputFile, outputContent)
	batch.OutputFileID = &outputFile.ID

	store.SaveBatch(batch)
	c.JSON(http.StatusOK, batch)
}

// HandleListBatches lists batches.
//
// @Summary		List batches
// @Tags			Batch
// @Produce		json
// @Success		200	{object}	models.ListBatchesResponse
// @Failure		401	{object}	errs.Response
// @Security		BearerAuth
// @Router			/batches [get]
func HandleListBatches(c *gin.Context) {
	batches := store.ListBatches()
	firstID, lastID := "", ""
	if len(batches) > 0 {
		firstID, lastID = batches[0].ID, batches[len(batches)-1].ID
	}
	c.JSON(http.StatusOK, gin.H{
		"object":   "list",
		"data":     batches,
		"first_id": firstID,
		"last_id":  lastID,
		"has_more": false,
	})
}

// HandleGetBatch retrieves a batch by id.
//
// @Summary		Get batch
// @Tags			Batch
// @Produce		json
// @Param			batch_id	path		string	true	"Batch ID"
// @Success		200			{object}	models.Batch
// @Failure		404			{object}	errs.Response
// @Security		BearerAuth
// @Router			/batches/{batch_id} [get]
func HandleGetBatch(c *gin.Context) {
	id := c.Param("batch_id")
	batch, ok := store.GetBatch(id)
	if !ok {
		errs.NotFound(c, fmt.Sprintf("No such batch: '%s'", id), "batch_not_found")
		return
	}
	c.JSON(http.StatusOK, batch)
}

// HandleCancelBatch cancels an in-progress batch.
//
// @Summary		Cancel batch
// @Description	Cancels an in-progress batch. In this mock, batches already complete synchronously on creation, so the batch is returned unchanged.
// @Tags			Batch
// @Produce		json
// @Param			batch_id	path		string	true	"Batch ID"
// @Success		200			{object}	models.Batch
// @Failure		404			{object}	errs.Response
// @Security		BearerAuth
// @Router			/batches/{batch_id}/cancel [post]
func HandleCancelBatch(c *gin.Context) {
	id := c.Param("batch_id")
	batch, ok := store.GetBatch(id)
	if !ok {
		errs.NotFound(c, fmt.Sprintf("No such batch: '%s'", id), "batch_not_found")
		return
	}
	// Batches complete synchronously on creation in this mock, so by the time
	// a cancel request can arrive the batch is already "completed" — return
	// it unchanged rather than simulating a cancelling/cancelled transition.
	c.JSON(http.StatusOK, batch)
}
