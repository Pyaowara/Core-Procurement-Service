package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type POStatus string

const (
	POStatusDraft     POStatus = "DRAFT"
	POStatusSent      POStatus = "SENT"
	POStatusCompleted POStatus = "COMPLETED"
)

type PurchaseOrder struct {
	gorm.Model
	ID             uint     `gorm:"primaryKey"`
	PONumber       string   `gorm:"unique;not null"`
	PRID           uint     `gorm:"not null;index"`
	VendorID       uint     `gorm:"not null;index"`
	Status         POStatus `gorm:"type:varchar(20);default:'DRAFT'"`
	CreditDay      int      `gorm:"default:0"`
	DueDate        time.Time
	Items          []POItem        `gorm:"foreignKey:POID"`
	VendorSnapshot *VendorSnapshot `gorm:"foreignKey:POID"`
	IsDeleted      bool            `gorm:"default:false"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

type POItem struct {
	gorm.Model
	ID           uint   `gorm:"primaryKey"`
	POID         uint   `gorm:"not null;index"`
	ItemName     string `gorm:"not null"`
	Description  string
	Quantity     int     `gorm:"not null"`
	Unit         string  `gorm:"not null"`
	PricePerUnit float64 `gorm:"type:decimal(15,2);not null"`
	Discount     float64 `gorm:"type:decimal(15,2)"`
	DiscountUnit string  `gorm:"type:varchar(10)"` // % or BAHT
	TotalPrice   float64 `gorm:"type:decimal(15,2);not null"`
	RequiredDate time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// GoodsReceived records when items are received for data tracking
type GoodsReceived struct {
	gorm.Model
	ID           uint            `gorm:"primaryKey"`
	POID         uint            `gorm:"not null;index"`
	ReceivedData json.RawMessage `gorm:"type:jsonb"`
	ReceivedAt   time.Time
	CreatedAt    time.Time
}
