package main

import (
	"log"
	"os"

	"github.com/core-procurement/purchase-service/config"
	"github.com/core-procurement/purchase-service/routes"
)

func main() {
	config.ConnectDatabase()

	r := routes.SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "6769"
	}

	log.Printf("purchase-service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
