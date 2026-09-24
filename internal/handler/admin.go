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

type AdminHandler struct {
	db *gorm.DB
}

func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

func (h *AdminHandler) Items(c *gin.Context) {
	page, pageSize, offset := parsePagination(c)
	query := h.db.Model(&model.Item{})
	if status := c.Query("status"); status != "" {
		if !validItemStatus(status) {
			response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
			return
		}
		query = query.Where("status = ?", status)
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
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

type itemReviewRequest struct {
	Status string `json:"status"`
	Remark string `json:"remark"`
}

func (h *AdminHandler) ReviewItem(c *gin.Context) {
	var req itemReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil || (req.Status != model.ItemStatusApproved && req.Status != model.ItemStatusClosed) {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var item model.Item
	if err := h.db.First(&item, c.Param("id")).Error; err != nil {
		respondDBError(c, err)
		return
	}
	if item.Status != model.ItemStatusPending {
		response.Error(c, http.StatusConflict, response.CodeInvalidState, "资源状态不允许当前操作")
		return
	}
	item.Status = req.Status
	item.Remark = strings.TrimSpace(req.Remark)
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&item).Error; err != nil {
			return err
		}
		if req.Status == model.ItemStatusClosed {
			return tx.Model(&model.Claim{}).
				Where("item_id = ? AND status = ?", item.ID, model.ClaimStatusPending).
				Updates(map[string]interface{}{
					"status": model.ClaimStatusRejected,
					"remark": "物品信息未通过审核",
				}).Error
		}
		return nil
	}); err != nil {
		respondDBError(c, err)
		return
	}
	h.db.Preload("User").First(&item, item.ID)
	response.Success(c, itemToView(item))
}

func (h *AdminHandler) Claims(c *gin.Context) {
	page, pageSize, offset := parsePagination(c)
	query := h.db.Model(&model.Claim{})
	if status := c.Query("status"); status != "" {
		if !validClaimStatus(status) {
			response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
			return
		}
		query = query.Where("claims.status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		respondDBError(c, err)
		return
	}
	var claims []model.Claim
	if err := query.Preload("User").Preload("Item").Preload("Item.User").
		Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&claims).Error; err != nil {
		respondDBError(c, err)
		return
	}
	list := make([]claimView, 0, len(claims))
	for _, claim := range claims {
		list = append(list, claimToView(claim))
	}
	response.Success(c, pageResult{List: list, Total: total, Page: page, PageSize: pageSize})
}

type claimReviewRequest struct {
	Status string `json:"status"`
	Remark string `json:"remark"`
}

func (h *AdminHandler) ReviewClaim(c *gin.Context) {
	var req claimReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil ||
		(req.Status != model.ClaimStatusApproved && req.Status != model.ClaimStatusRejected) {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var claim model.Claim
	if err := h.db.Preload("Item").First(&claim, c.Param("id")).Error; err != nil {
		respondDBError(c, err)
		return
	}
	if claim.Status != model.ClaimStatusPending || claim.Item == nil {
		response.Error(c, http.StatusConflict, response.CodeInvalidState, "资源状态不允许当前操作")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if req.Status == model.ClaimStatusApproved {
			if claim.Item.Status != model.ItemStatusApproved {
				return gorm.ErrInvalidData
			}
			if err := tx.Model(&model.Claim{}).
				Where("item_id = ? AND id <> ? AND status = ?", claim.ItemID, claim.ID, model.ClaimStatusPending).
				Updates(map[string]interface{}{
					"status": model.ClaimStatusRejected,
					"remark": "该物品已被其他申请认领",
				}).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.Item{}).Where("id = ?", claim.ItemID).
				Updates(map[string]interface{}{"status": model.ItemStatusClaimed}).Error; err != nil {
				return err
			}
		}
		return tx.Model(&model.Claim{}).Where("id = ?", claim.ID).
			Updates(map[string]interface{}{
				"status": req.Status,
				"remark": strings.TrimSpace(req.Remark),
			}).Error
	}); err != nil {
		if err == gorm.ErrInvalidData {
			response.Error(c, http.StatusConflict, response.CodeInvalidState, "资源状态不允许当前操作")
			return
		}
		respondDBError(c, err)
		return
	}

	h.db.Preload("User").Preload("Item").Preload("Item.User").First(&claim, claim.ID)
	response.Success(c, claimToView(claim))
}

func (h *AdminHandler) Users(c *gin.Context) {
	page, pageSize, offset := parsePagination(c)
	query := h.db.Model(&model.User{})
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		if uid, err := parseUint(keyword); err == nil {
			query = query.Where("uid = ? OR username LIKE ?", uid, "%"+keyword+"%")
		} else {
			query = query.Where("username LIKE ?", "%"+keyword+"%")
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		respondDBError(c, err)
		return
	}
	var users []model.User
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		respondDBError(c, err)
		return
	}
	list := make([]gin.H, 0, len(users))
	for _, user := range users {
		list = append(list, userToView(user))
	}
	response.Success(c, pageResult{List: list, Total: total, Page: page, PageSize: pageSize})
}

type userUpdateRequest struct {
	Role   *string `json:"role"`
	Status *string `json:"status"`
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	uid, ok := middleware.CurrentUID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录或令牌无效")
		return
	}
	targetUID, err := parseUint(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}
	if targetUID == uid {
		response.Error(c, http.StatusConflict, response.CodeInvalidState, "不能修改当前登录用户")
		return
	}

	var req userUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil || (req.Role == nil && req.Status == nil) {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}
	if req.Role != nil && !validRole(*req.Role) {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}
	if req.Status != nil && *req.Status != model.StatusActive && *req.Status != model.StatusDisabled {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var user model.User
	if err := h.db.First(&user, "uid = ?", targetUID).Error; err != nil {
		respondDBError(c, err)
		return
	}
	updates := make(map[string]interface{})
	if req.Role != nil {
		updates["role"] = *req.Role
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if err := h.db.Model(&user).Updates(updates).Error; err != nil {
		respondDBError(c, err)
		return
	}
	h.db.First(&user, "uid = ?", targetUID)
	response.Success(c, userToView(user))
}

func (h *AdminHandler) Statistics(c *gin.Context) {
	var stats struct {
		TotalItems   int64 `json:"total_items"`
		PendingItems int64 `json:"pending_items"`
		ClaimedItems int64 `json:"claimed_items"`
		TotalUsers   int64 `json:"total_users"`
		TotalClaims  int64 `json:"total_claims"`
	}
	if err := h.db.Model(&model.Item{}).Count(&stats.TotalItems).Error; err != nil {
		respondDBError(c, err)
		return
	}
	if err := h.db.Model(&model.Item{}).Where("status = ?", model.ItemStatusPending).Count(&stats.PendingItems).Error; err != nil {
		respondDBError(c, err)
		return
	}
	if err := h.db.Model(&model.Item{}).Where("status = ?", model.ItemStatusClaimed).Count(&stats.ClaimedItems).Error; err != nil {
		respondDBError(c, err)
		return
	}
	if err := h.db.Model(&model.User{}).Count(&stats.TotalUsers).Error; err != nil {
		respondDBError(c, err)
		return
	}
	if err := h.db.Model(&model.Claim{}).Count(&stats.TotalClaims).Error; err != nil {
		respondDBError(c, err)
		return
	}
	response.Success(c, stats)
}

func validRole(role string) bool {
	switch role {
	case model.RoleUser, model.RoleItemAdmin, model.RoleSystemAdmin:
		return true
	default:
		return false
	}
}
