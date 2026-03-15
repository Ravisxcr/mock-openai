package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"mock-openai/internal/mockdata" // Update this path
	"mock-openai/internal/models"   // Update this path
)

func HandleEmbeddings(c *gin.Context) {
	var req models.EmbeddingRequest

	// Bind the incoming JSON to our struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Build the OpenAI-compatible response
	resp := mockdata.GenerateMockEmbeddingResponse(req.Model, req.Input)

	// Send the JSON response with a 200 OK status
	c.JSON(http.StatusOK, resp)
}
