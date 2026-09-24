package model

import "time"

type Announcement struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"size:200;not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Published bool      `gorm:"not null;default:true" json:"published"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Announcement) TableName() string {
	return "announcements"
}
