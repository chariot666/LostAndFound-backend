package model

import "time"

const (
	ClaimStatusPending  = "pending"
	ClaimStatusApproved = "approved"
	ClaimStatusRejected = "rejected"
)

type Claim struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ItemID    uint      `gorm:"index;not null" json:"item_id"`
	UID       uint64    `gorm:"index;not null" json:"uid"`
	Proof     string    `gorm:"type:text;not null" json:"proof"`
	Contact   string    `gorm:"size:100;not null" json:"contact"`
	Status    string    `gorm:"size:20;index;not null" json:"status"`
	Remark    string    `gorm:"size:500" json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Item *Item `gorm:"foreignKey:ItemID" json:"-"`
	User *User `gorm:"foreignKey:UID;references:UID" json:"-"`
}

func (Claim) TableName() string {
	return "claims"
}
