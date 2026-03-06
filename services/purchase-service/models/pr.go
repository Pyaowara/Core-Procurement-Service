package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type PRStatus string

const (
	PRStatusDraft    PRStatus = "DRAFT"
	PRStatusPending  PRStatus = "PENDING"
	PRStatusApproved PRStatus = "APPROVED"
	PRStatusRejected PRStatus = "REJECTED"
)

type PurchaseRequest struct {
	gorm.Model
	ID          uint     `gorm:"primaryKey"`
	PRNumber    string   `gorm:"unique;not null"`
	RequesterID uint     `gorm:"not null"`
	Department  string   `gorm:"not null"`
	Status      PRStatus `gorm:"type:varchar(20);default:'DRAFT'"`
	WorkflowID  string   `gorm:"type:varchar(100)"`
	Items       []PRItem `gorm:"foreignKey:PRID"`
	IsDeleted   bool     `gorm:"default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type PRItem struct {
	gorm.Model
	ID                   uint   `gorm:"primaryKey"`
	PRID                 uint   `gorm:"not null;index"`
	ItemName             string `gorm:"not null"`
	Description          string
	Quantity             int     `gorm:"not null"`
	Unit                 string  `gorm:"not null"` // ชิ้น, แท่ง, etc.
	PricePerUnit         float64 `gorm:"type:decimal(15,2)"`
	Discount             float64 `gorm:"type:decimal(15,2)"`
	DiscountUnit         string  `gorm:"type:varchar(10)"` // % or BAHT
	TotalPrice           float64 `gorm:"type:decimal(15,2)"`
	RequiredDate         time.Time
	CurrentStockAtSubmit int       `gorm:"default:0"` // Stock snapshot at PR submission time
	StockCheckAt         time.Time // When stock was checked
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// InventorySnapshot stores the state of inventory at the time of PR submission for data consistency
type InventorySnapshot struct {
	gorm.Model
	ID           uint            `gorm:"primaryKey"`
	PRID         uint            `gorm:"not null;index;unique"`
	SnapshotData json.RawMessage `gorm:"type:jsonb"`
	CreatedAt    time.Time
}
