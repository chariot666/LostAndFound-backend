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

// List 我的通知列表: GET /api/v1/me/notifications
func (h *NotificationHandler) List(c *gin.Context) {
	uid, ok := middleware.CurrentUID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录或令牌无效")
		return
	}

	page, pageSize, offset := parsePagination(c)
	var total int64
	h.db.Model(&model.Notification{}).Where("uid = ?", uid).Count(&total)

	var list []model.Notification
	h.db.Where("uid = ?", uid).
		Order("created_at DESC").Offset(offset).Limit(pageSize).
		Find(&list)
	response.Success(c, pageResult{List: list, Total: total, Page: page, PageSize: pageSize})
}

// UnreadCount 未读通知数: GET /api/v1/me/notifications/unread-count
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	uid, ok := middleware.CurrentUID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录或令牌无效")
		return
	}
	var n int64
	h.db.Model(&model.Notification{}).Where("uid = ? AND is_read = ?", uid, false).Count(&n)
	response.Success(c, gin.H{"unread": n})
}

// ReadAll 全部标记已读: POST /api/v1/me/notifications/read-all
func (h *NotificationHandler) ReadAll(c *gin.Context) {
	uid, ok := middleware.CurrentUID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录或令牌无效")
		return
	}
	h.db.Model(&model.Notification{}).
		Where("uid = ? AND is_read = ?", uid, false).
		Update("is_read", true)
	response.Success(c, gin.H{"unread": 0})
}

// ReadOne 单条标记已读: POST /api/v1/me/notifications/:id/read
func (h *NotificationHandler) ReadOne(c *gin.Context) {
	uid, ok := middleware.CurrentUID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录或令牌无效")
		return
	}
	h.db.Model(&model.Notification{}).
		Where("id = ? AND uid = ?", c.Param("id"), uid).
		Update("is_read", true)
	response.Success(c, nil)
}
