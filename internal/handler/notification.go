package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"lost-found-server/internal/middleware"
	"lost-found-server/internal/model"
	"lost-found-server/internal/response"
)

type NotificationHandler struct {
	db *gorm.DB
}

func NewNotificationHandler(db *gorm.DB) *NotificationHandler {
	return &NotificationHandler{db: db}
}

func createNotification(db *gorm.DB, uid uint64, typ, title, content string, itemID uint) error {
	return db.Create(&model.Notification{
		UID:     uid,
		Type:    typ,
		Title:   title,
		Content: content,
		ItemID:  itemID,
		IsRead:  false,
	}).Error
}

func (h *NotificationHandler) List(c *gin.Context) {
	uid, ok := middleware.CurrentUID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录或令牌无效")
		return
	}

	page, pageSize, offset := parsePagination(c)
	query := h.db.Model(&model.Notification{}).Where("uid = ?", uid)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		respondDBError(c, err)
		return
	}

	var list []model.Notification
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		respondDBError(c, err)
		return
	}
	response.Success(c, pageResult{List: list, Total: total, Page: page, PageSize: pageSize})
}

func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	uid, ok := middleware.CurrentUID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录或令牌无效")
		return
	}
	var count int64
	if err := h.db.Model(&model.Notification{}).
		Where("uid = ? AND is_read = ?", uid, false).
		Count(&count).Error; err != nil {
		respondDBError(c, err)
		return
	}
	response.Success(c, gin.H{"unread": count})
}

func (h *NotificationHandler) ReadAll(c *gin.Context) {
	uid, ok := middleware.CurrentUID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录或令牌无效")
		return
	}
	if err := h.db.Model(&model.Notification{}).
		Where("uid = ? AND is_read = ?", uid, false).
		Update("is_read", true).Error; err != nil {
		respondDBError(c, err)
		return
	}
	response.Success(c, gin.H{"unread": 0})
}

func (h *NotificationHandler) ReadOne(c *gin.Context) {
	uid, ok := middleware.CurrentUID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录或令牌无效")
		return
	}
	notificationID, err := parseUint(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}
	var notification model.Notification
	if err := h.db.Where("id = ? AND uid = ?", notificationID, uid).First(&notification).Error; err != nil {
		respondDBError(c, err)
		return
	}
	if !notification.IsRead {
		if err := h.db.Model(&notification).Update("is_read", true).Error; err != nil {
			respondDBError(c, err)
			return
		}
	}
	response.Success(c, nil)
}
