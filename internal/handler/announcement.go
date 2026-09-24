package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"lost-found-server/internal/model"
	"lost-found-server/internal/response"
)

type AnnouncementHandler struct {
	db *gorm.DB
}

func NewAnnouncementHandler(db *gorm.DB) *AnnouncementHandler {
	return &AnnouncementHandler{db: db}
}

func (h *AnnouncementHandler) List(c *gin.Context) {
	h.list(c, true)
}

func (h *AnnouncementHandler) AdminList(c *gin.Context) {
	h.list(c, false)
}

func (h *AnnouncementHandler) list(c *gin.Context, publishedOnly bool) {
	page, pageSize, offset := parsePagination(c)
	query := h.db.Model(&model.Announcement{})
	if publishedOnly {
		query = query.Where("published = ?", true)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		respondDBError(c, err)
		return
	}
	var announcements []model.Announcement
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&announcements).Error; err != nil {
		respondDBError(c, err)
		return
	}
	response.Success(c, pageResult{List: announcements, Total: total, Page: page, PageSize: pageSize})
}

type announcementRequest struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	Published *bool  `json:"published"`
}

func validateAnnouncement(title, content string) bool {
	return strings.TrimSpace(title) != "" && len([]rune(title)) <= 100 &&
		strings.TrimSpace(content) != "" && len([]rune(content)) <= 10000
}

func (h *AnnouncementHandler) Create(c *gin.Context) {
	var req announcementRequest
	if err := c.ShouldBindJSON(&req); err != nil || !validateAnnouncement(req.Title, req.Content) {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}
	published := true
	if req.Published != nil {
		published = *req.Published
	}
	announcement := model.Announcement{
		Title:     strings.TrimSpace(req.Title),
		Content:   strings.TrimSpace(req.Content),
		Published: published,
	}
	if err := h.db.Create(&announcement).Error; err != nil {
		respondDBError(c, err)
		return
	}
	response.Success(c, announcement)
}

func (h *AnnouncementHandler) Update(c *gin.Context) {
	var req announcementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}
	var announcement model.Announcement
	if err := h.db.First(&announcement, c.Param("id")).Error; err != nil {
		respondDBError(c, err)
		return
	}
	if req.Title != "" {
		announcement.Title = strings.TrimSpace(req.Title)
	}
	if req.Content != "" {
		announcement.Content = strings.TrimSpace(req.Content)
	}
	if req.Published != nil {
		announcement.Published = *req.Published
	}
	if !validateAnnouncement(announcement.Title, announcement.Content) {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}
	if err := h.db.Save(&announcement).Error; err != nil {
		respondDBError(c, err)
		return
	}
	response.Success(c, announcement)
}

func (h *AnnouncementHandler) Delete(c *gin.Context) {
	var announcement model.Announcement
	if err := h.db.First(&announcement, c.Param("id")).Error; err != nil {
		respondDBError(c, err)
		return
	}
	if err := h.db.Delete(&announcement).Error; err != nil {
		respondDBError(c, err)
		return
	}
	response.Success(c, nil)
}
