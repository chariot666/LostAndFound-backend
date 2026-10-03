package model

import "time"

type Announcement struct {
	ID        uint      `gorm:"primaryKey;index:idx_announcements_cursor,priority:2" json:"id"`
	Title     string    `gorm:"size:200;not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Published bool      `gorm:"not null;default:true" json:"published"`
	CreatedAt time.Time `gorm:"index:idx_announcements_cursor,priority:1" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Announcement) TableName() string {
	return "announcements"
}
