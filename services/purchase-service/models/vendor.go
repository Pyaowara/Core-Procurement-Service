package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Vendor struct {
	gorm.Model
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	Address   string
	TaxID     string `gorm:"unique"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// VendorSnapshot stores vendor information at the time of PO creation for data consistency
type VendorSnapshot struct {
	gorm.Model
	ID            uint   `gorm:"primaryKey"`
	POID          uint   `gorm:"not null;index;unique"`
	VendorID      uint   `gorm:"not null;index"`
	VendorName    string `gorm:"not null"`
	VendorAddress string
	VendorTaxID   string
	SnapshotData  json.RawMessage `gorm:"type:jsonb"`
	CreatedAt     time.Time
}
