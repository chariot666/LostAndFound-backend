package model

import "time"

const (
	ReportStatusPending  = "pending"
	ReportStatusResolved = "resolved"
	ReportStatusRejected = "rejected"
)

type Report struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ItemID    uint      `gorm:"index;not null" json:"item_id"`
	UID       uint64    `gorm:"index;not null" json:"uid"`
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
