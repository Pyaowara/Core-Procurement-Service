package main

import (
	"log"
	"os"

	"github.com/core-procurement/approval-service/config"
	"github.com/core-procurement/approval-service/routes"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	config.ConnectDatabase()

	r := routes.SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "6770"
	}

	log.Printf("approval-service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
