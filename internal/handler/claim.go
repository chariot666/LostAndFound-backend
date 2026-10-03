package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"lost-found-server/internal/middleware"
	"lost-found-server/internal/model"
	"lost-found-server/internal/response"
)

type ClaimHandler struct {
	db *gorm.DB
}

func NewClaimHandler(db *gorm.DB) *ClaimHandler {
	return &ClaimHandler{db: db}
}

type claimRequest struct {
	Proof   string `json:"proof"`
	Contact string `json:"contact"`
}

func (h *ClaimHandler) Create(c *gin.Context) {
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
	if !canClaimItem(item.Type) {
		response.Error(c, http.StatusConflict, response.CodeInvalidState, "失物信息不能提交认领申请，请联系发布者")
		return
	}
	if item.Status != model.ItemStatusApproved {
		response.Error(c, http.StatusConflict, response.CodeInvalidState, "资源状态不允许当前操作")
		return
	}
	if item.UID == uid {
		response.Error(c, http.StatusForbidden, response.CodeForbidden, "不能认领自己发布的信息")
		return
	}

	var req claimRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Proof) == "" || strings.TrimSpace(req.Contact) == "" {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var existing model.Claim
	err := h.db.Where("item_id = ? AND uid = ? AND status IN ?", item.ID, uid,
		[]string{model.ClaimStatusPending, model.ClaimStatusApproved}).First(&existing).Error
	if err == nil {
		response.Error(c, http.StatusConflict, response.CodeDuplicateRequest, "请勿重复提交")
		return
	}
	if err != gorm.ErrRecordNotFound {
		respondDBError(c, err)
		return
	}

	claim := model.Claim{
		ItemID:  item.ID,
		UID:     uid,
		Proof:   strings.TrimSpace(req.Proof),
		Contact: strings.TrimSpace(req.Contact),
		Status:  model.ClaimStatusPending,
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&claim).Error; err != nil {
			return err
		}
		return createNotification(
			tx,
			item.UID,
			model.NotifTypeClaimSubmitted,
			"收到新的认领申请",
			"有人提交了你发布的拾物信息的认领申请，请前往后台查看。",
			item.ID,
		)
	}); err != nil {
		respondDBError(c, err)
		return
	}
	response.Success(c, gin.H{"id": claim.ID})
}

func canClaimItem(itemType string) bool {
	return itemType == model.ItemTypeFound
}

func (h *ClaimHandler) Mine(c *gin.Context) {
	uid, ok := middleware.CurrentUID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录或令牌无效")
		return
	}

	cursor, pageSize, err := parseCursorPagination(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "分页参数错误")
		return
	}
	query := h.db.Model(&model.Claim{}).Where("claims.uid = ?", uid)
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
	query = applyCursor(query, cursor)
	var claims []model.Claim
	if err := query.Preload("User").Preload("Item").Preload("Item.User").
		Order("created_at DESC, id DESC").Limit(pageSize + 1).Find(&claims).Error; err != nil {
		respondDBError(c, err)
		return
	}
	list := make([]claimView, 0, len(claims))
	for _, claim := range claims {
		list = append(list, claimToView(claim))
	}
	response.Success(c, makePageResult(list, total, pageSize, func(claim claimView) (time.Time, uint) {
		return claim.CreatedAt, claim.ID
	}))
}

func validClaimStatus(status string) bool {
	switch status {
	case model.ClaimStatusPending, model.ClaimStatusApproved, model.ClaimStatusRejected:
		return true
	default:
		return false
	}
}
