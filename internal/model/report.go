package model

import "time"

// 举报状态常量
const (
	ReportStatusPending  = "pending"  // 待处理
	ReportStatusResolved = "resolved" // 已处理（举报成立）
	ReportStatusRejected = "rejected" // 已驳回
)

// Report 举报表：用户举报不当物品
type Report struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ItemID    uint      `gorm:"index;not null" json:"item_id"`
	UID       uint64    `gorm:"index;not null" json:"uid"` // 举报人
	Reason    string    `gorm:"size:100;not null" json:"reason"`
	Detail    string    `gorm:"size:1000" json:"detail"`
	Status    string    `gorm:"size:20;index;not null;default:pending" json:"status"`
	Remark    string    `gorm:"size:500" json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Item *Item `gorm:"foreignKey:ItemID" json:"-"`
	User *User `gorm:"foreignKey:UID;references:UID" json:"-"`
}

func (Report) TableName() string {
	return "reports"
}
