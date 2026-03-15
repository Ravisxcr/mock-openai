package handlers

import (
	"github.com/gin-gonic/gin"
	"mock-openai/internal/mockdata"
	"mock-openai/internal/models"
	"net/http"
)

// POST /v1/images/generations
func HandleImageGeneration(c *gin.Context) {
	var req models.ImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	resp := mockdata.GenerateMockImageResponse("dall-e-3", req.Prompt)
	c.JSON(http.StatusOK, resp)
}

// POST /v1/images/edits
func HandleImageEdit(c *gin.Context) {
	// Mocking a successful edit response
	resp := mockdata.GenerateMockImageResponse("dall-e-2", "Edited version of uploaded image")
	c.JSON(http.StatusOK, resp)
}

// POST /v1/images/variations
func HandleImageVariation(c *gin.Context) {
	// Mocking a successful variation response
	resp := mockdata.GenerateMockImageResponse("dall-e-2", "Variation of uploaded image")
	c.JSON(http.StatusOK, resp)
}
