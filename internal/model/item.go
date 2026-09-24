package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	ItemTypeLost  = "lost"
	ItemTypeFound = "found"

	ItemStatusPending  = "pending"
	ItemStatusApproved = "approved"
	ItemStatusClaimed  = "claimed"
	ItemStatusClosed   = "closed"
)

type Item struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UID         uint64         `gorm:"index;not null" json:"uid"`
	Type        string         `gorm:"size:10;not null" json:"type"`
	Title       string         `gorm:"size:100;not null" json:"title"`
	Description string         `gorm:"type:text;not null" json:"description"`
	Location    string         `gorm:"size:100;not null" json:"location"`
	LostAt      time.Time      `gorm:"not null" json:"lost_at"`
	Contact     string         `gorm:"size:100;not null" json:"contact"`
	Images      []string       `gorm:"type:json;serializer:json" json:"images"`
	Status      string         `gorm:"size:20;index;not null" json:"status"`
	Remark      string         `gorm:"size:500" json:"remark"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	User *User `gorm:"foreignKey:UID;references:UID" json:"-"`
}

func (Item) TableName() string {
	return "items"
}
