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

type ReportHandler struct {
	db *gorm.DB
}

func NewReportHandler(db *gorm.DB) *ReportHandler {
	return &ReportHandler{db: db}
}

type reportRequest struct {
	Reason string `json:"reason"`
	Detail string `json:"detail"`
}

// Create 用户举报物品: POST /api/v1/items/:id/reports
func (h *ReportHandler) Create(c *gin.Context) {
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

	var req reportRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Reason) == "" {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "请填写举报原因")
		return
	}

	// 检查是否已举报过（待处理中的）
	var existing model.Report
	if err := h.db.Where("item_id = ? AND uid = ? AND status = ?",
		itemID, uid, model.ReportStatusPending).First(&existing).Error; err == nil {
		response.Error(c, http.StatusConflict, response.CodeDuplicateRequest, "您已举报过该物品，请等待处理")
		return
	} else if err != gorm.ErrRecordNotFound {
		respondDBError(c, err)
		return
	}

	report := model.Report{
		ItemID: uint(itemID),
		UID:    uid,
		Reason: strings.TrimSpace(req.Reason),
		Detail: strings.TrimSpace(req.Detail),
		Status: model.ReportStatusPending,
	}
	if err := h.db.Create(&report).Error; err != nil {
		respondDBError(c, err)
		return
	}
	response.Success(c, gin.H{"id": report.ID})
}

// AdminList 管理员看举报列表: GET /api/v1/admin/reports
func (h *ReportHandler) AdminList(c *gin.Context) {
	page, pageSize, offset := parsePagination(c)
	query := h.db.Model(&model.Report{})
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var reports []model.Report
	err := query.Preload("Item").Preload("Item.User").
		Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&reports).Error
	if err != nil {
		respondDBError(c, err)
		return
	}

	list := make([]gin.H, 0, len(reports))
	for _, r := range reports {
		entry := gin.H{
			"id": r.ID, "item_id": r.ItemID, "uid": r.UID,
			"reason": r.Reason, "detail": r.Detail,
			"status": r.Status, "remark": r.Remark,
			"created_at": r.CreatedAt,
		}
		if r.Item != nil {
			item := itemToView(*r.Item)
			entry["item"] = item
		}
		list = append(list, entry)
	}
	response.Success(c, pageResult{List: list, Total: total, Page: page, PageSize: pageSize})
}

type reportReviewRequest struct {
	Status string `json:"status"`
	Remark string `json:"remark"`
}

// Review 管理员处理举报: PUT /api/v1/admin/reports/:id
func (h *ReportHandler) Review(c *gin.Context) {
	var req reportReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil ||
		(req.Status != model.ReportStatusResolved && req.Status != model.ReportStatusRejected) {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var report model.Report
	if err := h.db.First(&report, c.Param("id")).Error; err != nil {
		respondDBError(c, err)
		return
	}
	if report.Status != model.ReportStatusPending {
		response.Error(c, http.StatusConflict, response.CodeInvalidState, "该举报已处理")
		return
	}

	report.Status = req.Status
	report.Remark = strings.TrimSpace(req.Remark)
	if err := h.db.Save(&report).Error; err != nil {
		respondDBError(c, err)
		return
	}

	// 举报成立则关闭物品
	if req.Status == model.ReportStatusResolved {
		h.db.Model(&model.Item{}).Where("id = ?", report.ItemID).
			Update("status", model.ItemStatusClosed)
	}
	response.Success(c, report)
}
