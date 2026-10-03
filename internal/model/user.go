package model

import "time"

const (
	RoleUser        = "user"
	RoleItemAdmin   = "item_admin"
	RoleSystemAdmin = "system_admin"

	StatusActive   = "active"
	StatusDisabled = "disabled"
)

type User struct {
	UID          uint64    `gorm:"primaryKey;autoIncrement;index:idx_users_cursor,priority:2" json:"uid"`
	Username     string    `gorm:"size:30;not null" json:"username"`
	Contact      string    `gorm:"size:100;not null;default:''" json:"contact"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Role         string    `gorm:"size:20;not null;default:user" json:"role"`
	Status       string    `gorm:"size:20;not null;default:active" json:"status"`
	CreatedAt    time.Time `gorm:"index:idx_users_cursor,priority:1" json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
