package models

import (
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
	Purpose     string   `gorm:"type:text"` // Purpose of the purchase request
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
	SKU                  string `gorm:"not null;index"` // Stock Keeping Unit
	ItemName             string `gorm:"not null"`
	Description          string
	Quantity             int     `gorm:"not null"`
	PricePerUnit         float64 `gorm:"type:decimal(15,2)"`
	Discount             float64 `gorm:"type:decimal(15,2)"`
	DiscountUnit         string  `gorm:"type:varchar(10)"` // % or BAHT
	TotalPrice           float64 `gorm:"type:decimal(15,2)"`
	RequiredDate         time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
