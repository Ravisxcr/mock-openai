package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"mock-openai/internal/errs"
	"mock-openai/internal/mockdata"
	"mock-openai/internal/models"
)

// HandleEmbeddings creates an embedding vector for the given input.
//
// @Summary		Create embeddings
// @Description	Creates an embedding vector representing the input text.
// @Tags			Embeddings
// @Accept			json
// @Produce		json
// @Param			request	body		models.EmbeddingRequest	true	"Embedding request"
// @Success		200		{object}	models.EmbeddingResponse
// @Failure		400		{object}	errs.Response
// @Failure		401		{object}	errs.Response
// @Security		BearerAuth
// @Router			/embeddings [post]
func HandleEmbeddings(c *gin.Context) {
	var req models.EmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errs.InvalidRequest(c, "Invalid request body: "+err.Error(), "")
		return
	}
	if req.Model == "" {
		errs.InvalidRequest(c, "you must provide a model parameter", "model")
		return
	}
	if req.Input == nil {
		errs.InvalidRequest(c, "you must provide an input parameter", "input")
		return
	}

	resp := mockdata.GenerateMockEmbeddingResponse(req)
	c.JSON(http.StatusOK, resp)
}
