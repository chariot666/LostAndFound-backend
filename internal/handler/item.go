package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"lost-found-server/internal/middleware"
	"lost-found-server/internal/model"
	"lost-found-server/internal/response"
)

type ItemHandler struct {
	db *gorm.DB
}

func NewItemHandler(db *gorm.DB) *ItemHandler {
	return &ItemHandler{db: db}
}

type itemRequest struct {
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Location    string   `json:"location"`
	LostAt      string   `json:"lost_at"`
	Contact     string   `json:"contact"`
	Images      []string `json:"images"`
}

type itemUpdateRequest struct {
	Type        *string   `json:"type"`
	Title       *string   `json:"title"`
	Description *string   `json:"description"`
	Location    *string   `json:"location"`
	LostAt      *string   `json:"lost_at"`
	Contact     *string   `json:"contact"`
	Images      *[]string `json:"images"`
}

func (h *ItemHandler) List(c *gin.Context) {
	page, pageSize, offset := parsePagination(c)
	query := h.db.Model(&model.Item{}).Where("status IN ?", []string{model.ItemStatusApproved, model.ItemStatusClaimed})

	if itemType := c.Query("type"); itemType != "" {
		if itemType != model.ItemTypeLost && itemType != model.ItemTypeFound {
			response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
			return
		}
		query = query.Where("type = ?", itemType)
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR description LIKE ?", like, like)
	}
	if location := strings.TrimSpace(c.Query("location")); location != "" {
		query = query.Where("location LIKE ?", "%"+location+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		respondDBError(c, err)
		return
	}

	var items []model.Item
	if err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		respondDBError(c, err)
		return
	}

	list := make([]itemView, 0, len(items))
	for _, item := range items {
		list = append(list, itemToView(item))
	}
	response.Success(c, pageResult{List: list, Total: total, Page: page, PageSize: pageSize})
}

func (h *ItemHandler) Detail(c *gin.Context) {
	var item model.Item
	if err := h.db.Preload("User").First(&item, c.Param("id")).Error; err != nil {
		respondDBError(c, err)
		return
	}

	if item.Status != model.ItemStatusApproved && item.Status != model.ItemStatusClaimed {
		uid, loggedIn := middleware.CurrentUID(c)
		role, _ := c.Get(middleware.ContextRole)
		isAdmin := role == model.RoleItemAdmin || role == model.RoleSystemAdmin
		if !loggedIn || (uid != item.UID && !isAdmin) {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, "资源不存在")
			return
		}
	}

	response.Success(c, itemToView(item))
}

func (h *ItemHandler) Create(c *gin.Context) {
	uid, ok := middleware.CurrentUID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录或令牌无效")
		return
	}

	var req itemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}
	lostAt, err := parseTimeValue(req.LostAt)
	if err != nil || !validateItemInput(req.Type, req.Title, req.Description, req.Location, req.Contact, req.Images) {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	item := model.Item{
		UID:         uid,
		Type:        strings.TrimSpace(req.Type),
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Location:    strings.TrimSpace(req.Location),
		LostAt:      lostAt,
		Contact:     strings.TrimSpace(req.Contact),
		Images:      req.Images,
		Status:      model.ItemStatusPending,
	}
	if err := h.db.Create(&item).Error; err != nil {
		respondDBError(c, err)
		return
	}
	h.db.Preload("User").First(&item, item.ID)
	response.Success(c, itemToView(item))
}

func (h *ItemHandler) Update(c *gin.Context) {
	uid, ok := middleware.CurrentUID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录或令牌无效")
		return
	}

	var item model.Item
	if err := h.db.First(&item, c.Param("id")).Error; err != nil {
		respondDBError(c, err)
		return
	}
	if item.UID != uid {
		response.Error(c, http.StatusForbidden, response.CodeForbidden, "没有权限")
		return
	}
	if item.Status == model.ItemStatusClaimed || item.Status == model.ItemStatusClosed {
		response.Error(c, http.StatusConflict, response.CodeInvalidState, "资源状态不允许当前操作")
		return
	}

	var req itemUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	next := item
	if req.Type != nil {
		next.Type = strings.TrimSpace(*req.Type)
	}
	if req.Title != nil {
		next.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		next.Description = strings.TrimSpace(*req.Description)
	}
	if req.Location != nil {
		next.Location = strings.TrimSpace(*req.Location)
	}
	if req.Contact != nil {
		next.Contact = strings.TrimSpace(*req.Contact)
	}
	if req.Images != nil {
		next.Images = *req.Images
	}
	if req.LostAt != nil {
		lostAt, parseErr := parseTimeValue(*req.LostAt)
		if parseErr != nil {
			response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
			return
		}
		next.LostAt = lostAt
	}
	if !validateItemInput(next.Type, next.Title, next.Description, next.Location, next.Contact, next.Images) {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	item.Type = next.Type
	item.Title = next.Title
	item.Description = next.Description
	item.Location = next.Location
	item.LostAt = next.LostAt
	item.Contact = next.Contact
	item.Images = next.Images
	item.Status = model.ItemStatusPending
	item.Remark = ""
	if err := h.db.Save(&item).Error; err != nil {
		respondDBError(c, err)
		return
	}
	h.db.Preload("User").First(&item, item.ID)
	response.Success(c, itemToView(item))
}

func (h *ItemHandler) Delete(c *gin.Context) {
	uid, ok := middleware.CurrentUID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录或令牌无效")
		return
	}

	var item model.Item
	if err := h.db.First(&item, c.Param("id")).Error; err != nil {
		respondDBError(c, err)
		return
	}
	if item.UID != uid {
		response.Error(c, http.StatusForbidden, response.CodeForbidden, "没有权限")
		return
	}
	if err := h.db.Delete(&item).Error; err != nil {
		respondDBError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *ItemHandler) MyItems(c *gin.Context) {
	uid, ok := middleware.CurrentUID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录或令牌无效")
		return
	}

	page, pageSize, offset := parsePagination(c)
	query := h.db.Model(&model.Item{}).Where("uid = ?", uid)
	if status := c.Query("status"); status != "" {
		if !validItemStatus(status) {
			response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
			return
		}
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		respondDBError(c, err)
		return
	}
	var items []model.Item
	if err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		respondDBError(c, err)
		return
	}
	list := make([]itemView, 0, len(items))
	for _, item := range items {
		list = append(list, itemToView(item))
	}
	response.Success(c, pageResult{List: list, Total: total, Page: page, PageSize: pageSize})
}

func validItemStatus(status string) bool {
	switch status {
	case model.ItemStatusPending, model.ItemStatusApproved, model.ItemStatusClaimed, model.ItemStatusClosed:
		return true
	default:
		return false
	}
}
