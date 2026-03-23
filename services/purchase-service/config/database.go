package config

import (
	"log"
	"os"

	"github.com/core-procurement/purchase-service/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN environment variable is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	log.Println("database connection established")

	// Use migrator to sync database schema with models
	migrator := db.Migrator()

	// AutoMigrate creates/updates tables based on models
	migrator.AutoMigrate(
		&models.PurchaseRequest{},
		&models.PRItem{},
		&models.InventorySnapshot{},
		&models.PurchaseOrder{},
		&models.POItem{},
		&models.GoodsReceived{},
		&models.Vendor{},
	)

	// Clean up InventorySnapshot table - remove columns not in current model
	// This ensures the table matches the model definition exactly
	if migrator.HasTable(&models.InventorySnapshot{}) {
		// Drop any extra columns that shouldn't be there
		if migrator.HasColumn(&models.InventorySnapshot{}, "sku") {
			log.Println("Removing SKU column from inventory_snapshots table...")
			migrator.DropColumn(&models.InventorySnapshot{}, "sku")
		}
	}

	DB = db
}
