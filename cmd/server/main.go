package main

import (
	"log"

	"mock-openai/internal/config"   // Update this path
	"mock-openai/internal/database" // Update this path
	"mock-openai/internal/router"   // Update this path
)

// @title						Mock OpenAI API
// @version					1.0
// @description				A lightweight local mock of the OpenAI REST API, covering Chat Completions
// @description				(including streaming, tool calls, and multimodal content), Embeddings, Images,
// @description				Audio, Models, Files, and Batch. Response shapes track the official OpenAI
// @description				OpenAPI spec vendored at openai.yml; see memory.md for known deviations.
// @termsOfService				http://swagger.io/terms/
//
// @license.name				MIT
//
// @host						localhost:8080
// @BasePath					/v1
//
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Bearer token auth. Any non-empty token is accepted unless API_KEY is configured (see README). Example: "Bearer sk-mock".
func main() {
	if err := config.LoadDotEnv(".env"); err != nil {
		log.Printf("Warning: failed to load .env file: %v", err)
	}

	apiKey := config.APIKey()
	if apiKey == "" {
		log.Println("No Auth: API_KEY not set in .env or environment — server will accept any Bearer token.")
	} else {
		log.Println("Auth enabled: requests must provide the configured API_KEY as a Bearer token.")
	}

	database.InitDB("mock_openai.db")
	// 3. Ensure the DB connection closes when the program exits
	defer database.CloseDB()

	// 1. Initialize the router
	r := router.SetupRouter(apiKey)

	// 2. Start the server
	log.Println("Starting OpenAI Mock Server on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
