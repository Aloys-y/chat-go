package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username     string    `gorm:"uniqueIndex;not null"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	DisplayName  string    `gorm:"not null"`
	LastLogin    time.Time
	IsOnline     bool      `gorm:"default:false"`
	Rooms        []*Room   `gorm:"many2many:user_rooms;"`
}