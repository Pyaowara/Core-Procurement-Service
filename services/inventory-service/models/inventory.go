package models

import "gorm.io/gorm"

type Inventory struct {
	gorm.Model
	Sku string `gorm:"unique;not null"`
	Name string `gorm:"not null"`
	Description string `gorm:"not null"`
	Quantity int `gorm:"not null"`
}
