package model

import "time"

// Favorite records which user has bookmarked which item.
type Favorite struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UID       uint64    `gorm:"index:idx_user_item,unique" json:"uid"`
	ItemID    uint      `gorm:"index:idx_user_item,unique" json:"item_id"`
	CreatedAt time.Time `json:"created_at"`

	Item *Item `gorm:"foreignKey:ItemID" json:"-"`
}

func (Favorite) TableName() string {
	return "favorites"
}
