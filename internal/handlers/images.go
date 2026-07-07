package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"mock-openai/internal/errs"
	"mock-openai/internal/mockdata"
	"mock-openai/internal/models"
)

// HandleImageGeneration creates an image from a text prompt.
//
// @Summary		Create image
// @Description	Creates an image given a text prompt.
// @Tags			Images
// @Accept			json
// @Produce		json
// @Param			request	body		models.ImageRequest	true	"Image generation request"
// @Success		200		{object}	models.ImageResponse
// @Failure		400		{object}	errs.Response
// @Failure		401		{object}	errs.Response
// @Security		BearerAuth
// @Router			/images/generations [post]
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

// HandleImageEdit creates an edited or extended image given an original image and a prompt.
//
// @Summary		Create image edit
// @Description	Creates an edited or extended image given an original image and a prompt.
// @Tags			Images
// @Accept			multipart/form-data
// @Produce		json
// @Param			image			formData	file	true	"Image to edit"
// @Param			prompt			formData	string	true	"Description of the desired edit"
// @Param			response_format	formData	string	false	"url or b64_json"
// @Success		200				{object}	models.ImageResponse
// @Failure		400				{object}	errs.Response
// @Failure		401				{object}	errs.Response
// @Security		BearerAuth
// @Router			/images/edits [post]
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

// HandleImageVariation creates a variation of a given image.
//
// @Summary		Create image variation
// @Description	Creates a variation of a given image.
// @Tags			Images
// @Accept			multipart/form-data
// @Produce		json
// @Param			image			formData	file	true	"Image to use as the basis for the variation"
// @Param			response_format	formData	string	false	"url or b64_json"
// @Success		200				{object}	models.ImageResponse
// @Failure		400				{object}	errs.Response
// @Failure		401				{object}	errs.Response
// @Security		BearerAuth
// @Router			/images/variations [post]
func HandleImageVariation(c *gin.Context) {
	if _, _, err := c.Request.FormFile("image"); err != nil {
		errs.InvalidRequest(c, "you must provide an image file", "image")
		return
	}

	resp := mockdata.GenerateMockImageResponse("dall-e-2", "Variation of uploaded image", c.PostForm("response_format"), 1)
	c.JSON(http.StatusOK, resp)
}
