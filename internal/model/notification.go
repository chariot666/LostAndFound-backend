package model

import "time"

// 通知类型常量
const (
	NotifTypeClaimSubmitted = "claim_submitted" // 有人认领了我的物品
	NotifTypeClaimApproved  = "claim_approved"   // 我的认领申请通过了
	NotifTypeClaimRejected  = "claim_rejected"   // 我的认领申请被拒绝
	NotifTypeItemApproved   = "item_approved"    // 我发布的物品审核通过
	NotifTypeItemRejected   = "item_rejected"    // 我发布的物品被拒绝
)

// Notification 站内通知表
type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UID       uint64    `gorm:"index;not null" json:"uid"` // 接收通知的用户
	Type      string    `gorm:"size:30;not null" json:"type"`
	Title     string    `gorm:"size:200;not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	ItemID    uint      `gorm:"index" json:"item_id"`
	IsRead    bool      `gorm:"not null;default:false" json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

func (Notification) TableName() string {
	return "notifications"
}
