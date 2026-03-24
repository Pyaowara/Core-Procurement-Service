package main

import (
	"log"
	"os"

	"github.com/core-procurement/inventory-service/config"
	"github.com/core-procurement/inventory-service/messaging"
	"github.com/core-procurement/inventory-service/routes"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	config.ConnectDatabase()

	// Connect to RabbitMQ and start inventory event consumer
	if err := messaging.ConnectRabbitMQ(); err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer messaging.MQClient.Close()

	if err := messaging.MQClient.DeclareExchange(messaging.ExchangeName); err != nil {
		log.Fatalf("failed to declare exchange: %v", err)
	}

	go messaging.StartGoodsReceivedConsumer()

	r := routes.SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "6768"
	}

	log.Printf("inventory-service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
