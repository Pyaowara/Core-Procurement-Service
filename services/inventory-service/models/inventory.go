package models

import "gorm.io/gorm"

type Inventory struct {
	gorm.Model
	Name string `gorm:"unique;not null"`
	Description string `gorm:"not null"`
	Quantity int `gorm:"not null"`
	UnitPrice float64 `gorm:"not null"`
}
