package model

import "time"

// Favorite 收藏表：记录哪个用户收藏了哪个物品
type Favorite struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UID       uint64    `gorm:"index:idx_user_item,unique" json:"uid"`        // 收藏者
	ItemID    uint      `gorm:"index:idx_user_item,unique" json:"item_id"`   // 被收藏的物品
	CreatedAt time.Time `json:"created_at"`

	Item *Item `gorm:"foreignKey:ItemID" json:"-"`
}

func (Favorite) TableName() string {
	return "favorites"
}
