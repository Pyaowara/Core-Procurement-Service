package main

import (
	"log"
	"os"

	"github.com/core-procurement/approval-service/config"
	"github.com/core-procurement/approval-service/messaging"
	"github.com/core-procurement/approval-service/routes"
	"github.com/core-procurement/approval-service/services"
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

	// Start event subscribers in a goroutine
	go services.SubscribeToPREvents()

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
