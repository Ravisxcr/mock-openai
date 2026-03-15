package router

import (
	"net/http"
	"strings"
	"time"

	"mock-openai/internal/database" // Import your db package

	"github.com/gin-gonic/gin"
)

// AuthMiddleware checks for a Bearer token in the Authorization header
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// 1. Check if header exists
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort() // Stops the request from reaching your handlers
			return
		}

		// 2. Check for "Bearer " prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be in 'Bearer <token>' format"})
			c.Abort()
			return
		}

		// 3. Validate the token
		// For a mock server, we can accept any token (like 'mock-key')
		// or check for a specific one.
		token := parts[1]
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// If everything is fine, call Next() to proceed to the handler
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
