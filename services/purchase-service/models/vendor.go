package models

import (
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
