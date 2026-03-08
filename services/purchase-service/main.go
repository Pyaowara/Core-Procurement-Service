package main

import (
	"log"
	"os"

	"github.com/core-procurement/purchase-service/config"
	"github.com/core-procurement/purchase-service/messaging"
	"github.com/core-procurement/purchase-service/routes"
	"github.com/core-procurement/purchase-service/services"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	config.ConnectDatabase()

	// Connect to RabbitMQ
	if err := messaging.ConnectRabbitMQ(); err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer messaging.MQClient.Close()

	// Setup message broker
	if err := messaging.MQClient.DeclareExchange(messaging.ExchangeName); err != nil {
		log.Fatalf("failed to declare exchange: %v", err)
	}

	// Start event subscribers in goroutines
	go services.SubscribeToApprovalEvents()
	go services.SubscribeToInventoryEvents()

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
