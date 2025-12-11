package models

import (
	"github.com/jinzhu/gorm"
)

type Room struct {
	gorm.Model
	Name        string   `gorm:"unique_index;not null"`
	Description string
	IsPublic    bool     `gorm:"default:true"`
	OwnerID     uint     `gorm:"not null"`
	Owner       *User    `gorm:"foreignkey:OwnerID"`
	Users       []*User  `gorm:"many2many:user_rooms;"`
}