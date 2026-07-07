package router

import (
	"strings"
	"time"

	"mock-openai/internal/apierror"
	"mock-openai/internal/database" // Import your db package

	"github.com/gin-gonic/gin"
)

// AuthMiddleware checks for a Bearer token in the Authorization header.
// Any non-empty token is accepted — this is a mock server, not a real
// key validator — but the presence/shape check matches real API behavior.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			apierror.Respond(c, 401, "invalid_request_error",
				"You didn't provide an API key. You need to provide your API key in an Authorization header using Bearer auth (i.e. Authorization: Bearer YOUR_KEY).",
				"", "")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
			apierror.Respond(c, 401, "invalid_request_error",
				"Authorization header must be in the format 'Bearer <token>'.", "", "")
			return
		}

		c.Next()
	}
}

func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start).String()

		// Capture data
		method := c.Request.Method
		path := c.Request.URL.Path
		ip := c.ClientIP()
		status := c.Writer.Status()

		// Save to SQLite (using a goroutine so it doesn't slow down the response)
		go database.SaveLog(method, path, ip, status, latency)
	}
}
