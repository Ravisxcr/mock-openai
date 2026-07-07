package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"mock-openai/internal/apierror"
	"mock-openai/internal/mockdata"
	"mock-openai/internal/models"
)

// POST /v1/embeddings
func HandleEmbeddings(c *gin.Context) {
	var req models.EmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidRequest(c, "Invalid request body: "+err.Error(), "")
		return
	}
	if req.Model == "" {
		apierror.InvalidRequest(c, "you must provide a model parameter", "model")
		return
	}
	if req.Input == nil {
		apierror.InvalidRequest(c, "you must provide an input parameter", "input")
		return
	}

	resp := mockdata.GenerateMockEmbeddingResponse(req)
	c.JSON(http.StatusOK, resp)
}
