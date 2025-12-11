package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Username     string    `gorm:"unique_index;not null"`
	Email        string    `gorm:"unique_index;not null"`
	PasswordHash string    `gorm:"not null"`
	DisplayName  string    `gorm:"not null"`
	LastLogin    time.Time
	IsOnline     bool      `gorm:"default:false"`
	Rooms        []*Room   `gorm:"many2many:user_rooms;"`
}