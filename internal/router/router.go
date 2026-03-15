package router

import (
	"github.com/gin-gonic/gin"
	"mock-openai/internal/handlers" // Update this path
)

// SetupRouter initializes the Gin engine and defines all routes.
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Version 1 API Group
	// API root
	api := r.Group("/v1")
		{
			// Chat routes
			api.POST("/completions", handlers.HandleLegacyCompletions)

			// Chat Completions Group
			chat := api.Group("/chat/completions")
			{
				chat.POST("", handlers.HandleChatCompletions)           // Create
				chat.GET("", handlers.HandleListChatCompletions)       // List
				
				// Specific ID routes
				chat.GET("/:completion_id", handlers.HandleChatCompletion)         // Get
				chat.POST("/:completion_id", handlers.HandleUpdateChatCompletion)   // Update
				chat.DELETE("/:completion_id", handlers.HandleDeleteChatCompletion) // Delete
				
				// Nested Resources
				chat.GET("/:completion_id/messages", handlers.HandleGetChatMessages)
			}
		}

		// Model routes
		models := api.Group("/models")
		{
			models.GET("", handlers.HandleModels)
			models.GET("/:model_id", handlers.HandleModel)
		}

		// --- EMBEDDINGS GROUP ---
		api.POST("/embeddings", handlers.HandleEmbeddings)

		// --- IMAGES GROUP ---
		images := api.Group("/images")
		{
			images.POST("/generations", handlers.HandleImageGeneration)
			images.POST("/edits", handlers.HandleImageEdit)
			images.POST("/variations", handlers.HandleImageVariation)
		}

		// --- AUDIO GROUP ---
		audio := api.Group("/audio")
		{
			audio.POST("/transcriptions", handlers.HandleAudioTranscription)
			audio.POST("/translations", handlers.HandleAudioTranslation)
			audio.POST("/speech", handlers.HandleAudioSpeech)
			
			// Voice Consents
			consents := audio.Group("/voice_consents")
			{
				consents.GET("", handlers.HandleListVoiceConsents)
				consents.POST("", handlers.HandleCreateVoiceConsent)
				consents.GET("/:consent_id", handlers.HandleGetVoiceConsent)
				consents.POST("/:consent_id", handlers.HandleCreateVoiceConsent) // Update is often a POST
				consents.DELETE("/:consent_id", handlers.HandleDeleteVoiceConsent)
			}
		}

	// Health check for the mock server itself
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "running"})
	})

	return r
}