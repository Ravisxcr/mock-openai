// Package errs provides the OpenAI-compatible error envelope
// (see openai.yml components.schemas.Error / ErrorResponse).
package errs

import "github.com/gin-gonic/gin"

// Detail mirrors the OpenAI `Error` schema.
type Detail struct {
	Message string  `json:"message"`
	Type    string  `json:"type"`
	Param   *string `json:"param"`
	Code    *string `json:"code"`
}

// Response mirrors the OpenAI `ErrorResponse` schema.
type Response struct {
	Error Detail `json:"error"`
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Respond writes an OpenAI-shaped error body and aborts the request.
func Respond(c *gin.Context, status int, errType, message, param, code string) {
	c.AbortWithStatusJSON(status, Response{
		Error: Detail{
			Message: message,
			Type:    errType,
			Param:   strPtr(param),
			Code:    strPtr(code),
		},
	})
}

// InvalidRequest is a convenience wrapper for the common 400 case.
func InvalidRequest(c *gin.Context, message, param string) {
	Respond(c, 400, "invalid_request_error", message, param, "")
}

// NotFound is a convenience wrapper for the common 404 case.
func NotFound(c *gin.Context, message, code string) {
	Respond(c, 404, "invalid_request_error", message, "", code)
}
