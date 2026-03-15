package main

import (
	"log"

	"mock-openai/internal/router" // Update this path
)

func main() {
	// 1. Initialize the router
	r := router.SetupRouter()

	// 2. Start the server
	log.Println("Starting OpenAI Mock Server on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}