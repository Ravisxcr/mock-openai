package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"mock-openai/internal/errs"
)

type modelInfo struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// knownModels is a static list of realistic current model ids, used by the
// models endpoints. Chat/embeddings/images/audio handlers themselves accept
// any model string — restricting those would hurt the mock's usefulness for
// testing arbitrary model names.
var knownModels = []modelInfo{
	{ID: "gpt-4o", Object: "model", Created: 1715367049, OwnedBy: "system"},
	{ID: "gpt-4o-mini", Object: "model", Created: 1721172741, OwnedBy: "system"},
	{ID: "gpt-4.1", Object: "model", Created: 1744651572, OwnedBy: "system"},
	{ID: "gpt-4.1-mini", Object: "model", Created: 1744651581, OwnedBy: "system"},
	{ID: "o1", Object: "model", Created: 1734375816, OwnedBy: "system"},
	{ID: "o3-mini", Object: "model", Created: 1737146383, OwnedBy: "system"},
	{ID: "gpt-3.5-turbo", Object: "model", Created: 1677610602, OwnedBy: "openai"},
	{ID: "text-embedding-3-small", Object: "model", Created: 1705948997, OwnedBy: "system"},
	{ID: "text-embedding-3-large", Object: "model", Created: 1705953180, OwnedBy: "system"},
	{ID: "text-embedding-ada-002", Object: "model", Created: 1671217299, OwnedBy: "openai-internal"},
	{ID: "dall-e-2", Object: "model", Created: 1698798177, OwnedBy: "system"},
	{ID: "dall-e-3", Object: "model", Created: 1698785189, OwnedBy: "system"},
	{ID: "gpt-image-1", Object: "model", Created: 1745517030, OwnedBy: "system"},
	{ID: "tts-1", Object: "model", Created: 1681940951, OwnedBy: "openai-internal"},
	{ID: "tts-1-hd", Object: "model", Created: 1699046015, OwnedBy: "system"},
	{ID: "whisper-1", Object: "model", Created: 1677532384, OwnedBy: "openai-internal"},
}

// HandleModels lists the available models.
//
// @Summary		List models
// @Description	Lists the model ids this mock recognizes. Note: chat/embeddings/images/audio endpoints accept any model string, not just these.
// @Tags			Models
// @Produce		json
// @Success		200	{object}	object{object=string,data=[]modelInfo}
// @Failure		401	{object}	errs.Response
// @Security		BearerAuth
// @Router			/models [get]
func HandleModels(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   knownModels,
	})
}

// HandleModel retrieves a single model by id.
//
// @Summary		Get model
// @Description	Retrieves a single model by id, if it's one of the ids returned by List models.
// @Tags			Models
// @Produce		json
// @Param			model_id	path		string	true	"Model ID"
// @Success		200			{object}	modelInfo
// @Failure		404			{object}	errs.Response
// @Security		BearerAuth
// @Router			/models/{model_id} [get]
func HandleModel(c *gin.Context) {
	id := c.Param("model_id")
	for _, m := range knownModels {
		if m.ID == id {
			c.JSON(http.StatusOK, m)
			return
		}
	}
	errs.NotFound(c, fmt.Sprintf("The model '%s' does not exist", id), "model_not_found")
}
