package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"mock-openai/internal/errs"
	"mock-openai/internal/mockdata"
	"mock-openai/internal/models"
)

// POST /v1/embeddings
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
