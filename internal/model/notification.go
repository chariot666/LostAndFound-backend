package model

import "time"

const (
	NotifTypeClaimSubmitted = "claim_submitted"
	NotifTypeClaimApproved  = "claim_approved"
	NotifTypeClaimRejected  = "claim_rejected"
	NotifTypeItemApproved   = "item_approved"
	NotifTypeItemRejected   = "item_rejected"
)

type Notification struct {
	ID        uint      `gorm:"primaryKey;index:idx_notifications_cursor,priority:2" json:"id"`
	UID       uint64    `gorm:"index;not null" json:"uid"`
	Type      string    `gorm:"size:30;not null" json:"type"`
	Title     string    `gorm:"size:200;not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	ItemID    uint      `gorm:"index" json:"item_id"`
	IsRead    bool      `gorm:"not null;default:false" json:"is_read"`
	CreatedAt time.Time `gorm:"index:idx_notifications_cursor,priority:1" json:"created_at"`
}

func (Notification) TableName() string {
	return "notifications"
}
