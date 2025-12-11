package models

import (
	"gorm.io/gorm"
)

type Room struct {
	gorm.Model
	Name        string   `gorm:"uniqueIndex;not null"`
	Description string
	IsPublic    bool     `gorm:"default:true"`
	OwnerID     uint     `gorm:"not null"`
	Owner       *User    `gorm:"foreignKey:OwnerID"`
	Users       []*User  `gorm:"many2many:user_rooms;"`
}