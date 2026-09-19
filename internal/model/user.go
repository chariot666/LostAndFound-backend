package model

import "time"

const (
	RoleUser = "user"

	StatusActive   = "active"
	StatusDisabled = "disabled"
)

type User struct {
	UID          uint64    `gorm:"primaryKey;autoIncrement" json:"uid"`
	Username     string    `gorm:"size:30;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Role         string    `gorm:"size:20;not null;default:user" json:"role"`
	Status       string    `gorm:"size:20;not null;default:active" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
