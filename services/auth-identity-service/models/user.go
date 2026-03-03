package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
	Role     string `gorm:"not null;default:'Employee'"`
	FirstName string `gorm:"not null"`
	LastName string `gorm:"not null"`
	Email    string `gorm:"unique;not null"`
}
