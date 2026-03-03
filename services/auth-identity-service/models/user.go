package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
	Role     string `gorm:"not null"`
	FirstName string `gorm:"not null"`
	LastName string `gorm:"not null"`
	Email    string `gorm:"unique;not null"`
}
