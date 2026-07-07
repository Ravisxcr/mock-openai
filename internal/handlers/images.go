package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"mock-openai/internal/errs"
	"mock-openai/internal/mockdata"
	"mock-openai/internal/models"
)

// POST /v1/images/generations
func HandleImageGeneration(c *gin.Context) {
	var req models.ImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errs.InvalidRequest(c, "Invalid request body: "+err.Error(), "")
		return
	}
	if req.Prompt == "" {
		errs.InvalidRequest(c, "you must provide a prompt parameter", "prompt")
		return
	}
	model := req.Model
	if model == "" {
		model = "dall-e-2"
	}

	resp := mockdata.GenerateMockImageResponse(model, req.Prompt, req.ResponseFormat, req.N)
	c.JSON(http.StatusOK, resp)
}

// POST /v1/images/edits (multipart/form-data: image, prompt, ...)
func HandleImageEdit(c *gin.Context) {
	prompt := c.PostForm("prompt")
	if prompt == "" {
		errs.InvalidRequest(c, "you must provide a prompt parameter", "prompt")
		return
	}
	if _, _, err := c.Request.FormFile("image"); err != nil {
		errs.InvalidRequest(c, "you must provide an image file", "image")
		return
	}

	resp := mockdata.GenerateMockImageResponse("dall-e-2", prompt, c.PostForm("response_format"), 1)
	c.JSON(http.StatusOK, resp)
}

// POST /v1/images/variations (multipart/form-data: image, ...)
func HandleImageVariation(c *gin.Context) {
	if _, _, err := c.Request.FormFile("image"); err != nil {
		errs.InvalidRequest(c, "you must provide an image file", "image")
		return
	}

	resp := mockdata.GenerateMockImageResponse("dall-e-2", "Variation of uploaded image", c.PostForm("response_format"), 1)
	c.JSON(http.StatusOK, resp)
}
