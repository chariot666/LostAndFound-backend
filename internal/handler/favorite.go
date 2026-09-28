package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"lost-found-server/internal/middleware"
	"lost-found-server/internal/model"
	"lost-found-server/internal/response"
)

type FavoriteHandler struct {
	db *gorm.DB
}

func NewFavoriteHandler(db *gorm.DB) *FavoriteHandler {
	return &FavoriteHandler{db: db}
}

// Add 收藏物品: POST /api/v1/items/:id/favorite
func (h *FavoriteHandler) Add(c *gin.Context) {
	uid, ok := middleware.CurrentUID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录或令牌无效")
		return
	}

	itemID, err := parseUint(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	// 确认物品存在
	var item model.Item
	if err := h.db.First(&item, itemID).Error; err != nil {
		respondDBError(c, err)
		return
	}

	// 已收藏就直接返回（幂等：重复点不报错）
	var existing model.Favorite
	if err := h.db.Where("uid = ? AND item_id = ?", uid, itemID).First(&existing).Error; err == nil {
		response.Success(c, gin.H{"favorited": true})
		return
	} else if err != gorm.ErrRecordNotFound {
		respondDBError(c, err)
		return
	}

	favorite := model.Favorite{UID: uid, ItemID: uint(itemID)}
	if err := h.db.Create(&favorite).Error; err != nil {
		respondDBError(c, err)
		return
	}
	response.Success(c, gin.H{"favorited": true, "id": favorite.ID})
}

// Remove 取消收藏: DELETE /api/v1/items/:id/favorite
func (h *FavoriteHandler) Remove(c *gin.Context) {
	uid, ok := middleware.CurrentUID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录或令牌无效")
		return
	}

	itemID, err := parseUint(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	h.db.Where("uid = ? AND item_id = ?", uid, itemID).Delete(&model.Favorite{})
	response.Success(c, gin.H{"favorited": false})
}

// Mine 我的收藏列表: GET /api/v1/me/favorites
func (h *FavoriteHandler) Mine(c *gin.Context) {
	uid, ok := middleware.CurrentUID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录或令牌无效")
		return
	}

	page, pageSize, offset := parsePagination(c)

	var total int64
	h.db.Model(&model.Favorite{}).Where("uid = ?", uid).Count(&total)

	var favorites []model.Favorite
	err := h.db.Where("uid = ?", uid).
		Preload("Item").Preload("Item.User").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&favorites).Error
	if err != nil {
		respondDBError(c, err)
		return
	}

	list := make([]itemView, 0, len(favorites))
	for _, f := range favorites {
		if f.Item != nil {
			list = append(list, itemToView(*f.Item))
		}
	}
	response.Success(c, pageResult{List: list, Total: total, Page: page, PageSize: pageSize})
}
