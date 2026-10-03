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

	var item model.Item
	if err := h.db.First(&item, itemID).Error; err != nil {
		respondDBError(c, err)
		return
	}

	var existing model.Favorite
	switch err := h.db.Where("uid = ? AND item_id = ?", uid, itemID).First(&existing).Error; err {
	case nil:
		response.Success(c, gin.H{"favorited": true, "id": existing.ID})
		return
	case gorm.ErrRecordNotFound:
	default:
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

	var item model.Item
	if err := h.db.First(&item, itemID).Error; err != nil {
		respondDBError(c, err)
		return
	}
	if err := h.db.Where("uid = ? AND item_id = ?", uid, itemID).Delete(&model.Favorite{}).Error; err != nil {
		respondDBError(c, err)
		return
	}
	response.Success(c, gin.H{"favorited": false})
}

func (h *FavoriteHandler) Status(c *gin.Context) {
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

	var item model.Item
	if err := h.db.First(&item, itemID).Error; err != nil {
		respondDBError(c, err)
		return
	}

	var favorite model.Favorite
	err = h.db.Where("uid = ? AND item_id = ?", uid, itemID).First(&favorite).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		respondDBError(c, err)
		return
	}
	response.Success(c, gin.H{"favorited": err == nil})
}

func (h *FavoriteHandler) Mine(c *gin.Context) {
	uid, ok := middleware.CurrentUID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录或令牌无效")
		return
	}

	page, pageSize, offset := parsePagination(c)
	query := h.db.Model(&model.Favorite{}).Where("uid = ?", uid)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		respondDBError(c, err)
		return
	}

	var favorites []model.Favorite
	if err := query.Preload("Item").Preload("Item.User").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&favorites).Error; err != nil {
		respondDBError(c, err)
		return
	}

	list := make([]itemView, 0, len(favorites))
	for _, favorite := range favorites {
		if favorite.Item != nil {
			list = append(list, itemToView(*favorite.Item))
		}
	}
	response.Success(c, pageResult{List: list, Total: total, Page: page, PageSize: pageSize})
}
