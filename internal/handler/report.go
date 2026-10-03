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
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "请填写举报原因")
		return
	}
	req.Reason = strings.TrimSpace(req.Reason)
	req.Detail = strings.TrimSpace(req.Detail)
	if req.Reason == "" || len([]rune(req.Reason)) > 100 || len([]rune(req.Detail)) > 1000 {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "举报内容不符合规范")
		return
	}
	if item.UID == uid {
		response.Error(c, http.StatusForbidden, response.CodeForbidden, "不能举报自己发布的信息")
		return
	}

	var existing model.Report
	switch err := h.db.Where("item_id = ? AND uid = ? AND status = ?",
		itemID, uid, model.ReportStatusPending).First(&existing).Error; err {
	case nil:
		response.Error(c, http.StatusConflict, response.CodeDuplicateRequest, "您已举报过该物品，请等待处理")
		return
	case gorm.ErrRecordNotFound:
	default:
		respondDBError(c, err)
		return
	}

	report := model.Report{
		ItemID: uint(itemID),
		UID:    uid,
		Reason: req.Reason,
		Detail: req.Detail,
		Status: model.ReportStatusPending,
	}
	if err := h.db.Create(&report).Error; err != nil {
		respondDBError(c, err)
		return
	}
	response.Success(c, gin.H{"id": report.ID})
}

func (h *ReportHandler) AdminList(c *gin.Context) {
	cursor, pageSize, err := parseCursorPagination(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "分页参数错误")
		return
	}
	query := h.db.Model(&model.Report{})
	if status := c.Query("status"); status != "" {
		if !validReportStatus(status) {
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
	query = applyCursor(query, cursor)

	var reports []model.Report
	if err := query.Preload("Item").Preload("Item.User").Preload("User").
		Order("created_at DESC, id DESC").Limit(pageSize + 1).Find(&reports).Error; err != nil {
		respondDBError(c, err)
		return
	}

	hasMore := len(reports) > pageSize
	if hasMore {
		reports = reports[:pageSize]
	}
	nextCursor := ""
	if hasMore {
		last := reports[len(reports)-1]
		nextCursor = encodeCursor(last.CreatedAt, last.ID)
	}
	list := make([]gin.H, 0, len(reports))
	for _, report := range reports {
		entry := gin.H{
			"id":         report.ID,
			"item_id":    report.ItemID,
			"uid":        report.UID,
			"reason":     report.Reason,
			"detail":     report.Detail,
			"status":     report.Status,
			"remark":     report.Remark,
			"created_at": report.CreatedAt,
		}
		if report.User != nil {
			entry["username"] = report.User.Username
		}
		if report.Item != nil {
			entry["item"] = itemToView(*report.Item)
		}
		list = append(list, entry)
	}
	response.Success(c, pageResultWithCursor(list, total, pageSize, hasMore, nextCursor))
}

type reportReviewRequest struct {
	Status string `json:"status"`
	Remark string `json:"remark"`
}

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

	remark := strings.TrimSpace(req.Remark)
	if len([]rune(remark)) > 500 {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "审核备注过长")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.Report{}).
			Where("id = ? AND status = ?", report.ID, model.ReportStatusPending).
			Updates(map[string]interface{}{"status": req.Status, "remark": remark})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrInvalidData
		}
		if req.Status == model.ReportStatusResolved {
			return tx.Model(&model.Item{}).
				Where("id = ?", report.ItemID).
				Update("status", model.ItemStatusClosed).Error
		}
		return nil
	}); err != nil {
		if err == gorm.ErrInvalidData {
			response.Error(c, http.StatusConflict, response.CodeInvalidState, "该举报已处理")
			return
		}
		respondDBError(c, err)
		return
	}

	report.Status = req.Status
	report.Remark = remark
	response.Success(c, report)
}

func validReportStatus(status string) bool {
	switch status {
	case model.ReportStatusPending, model.ReportStatusResolved, model.ReportStatusRejected:
		return true
	default:
		return false
	}
}
