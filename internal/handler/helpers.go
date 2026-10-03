package handler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"lost-found-server/internal/model"
	"lost-found-server/internal/response"
)

type pageResult struct {
	List       interface{} `json:"list"`
	Total      int64       `json:"total"`
	PageSize   int         `json:"page_size"`
	NextCursor string      `json:"next_cursor,omitempty"`
	HasMore    bool        `json:"has_more"`
}

type cursorPosition struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uint      `json:"id"`
}

type userBrief struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
}

type itemView struct {
	ID          uint      `json:"id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	LostAt      time.Time `json:"lost_at"`
	Contact     string    `json:"contact"`
	Images      []string  `json:"images"`
	Status      string    `json:"status"`
	UID         uint64    `json:"uid,omitempty"`
	User        userBrief `json:"user"`
	Remark      string    `json:"remark,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

type claimView struct {
	ID        uint      `json:"id"`
	ItemID    uint      `json:"item_id"`
	UID       uint64    `json:"uid"`
	Username  string    `json:"username"`
	Proof     string    `json:"proof"`
	Contact   string    `json:"contact"`
	Status    string    `json:"status"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	Item      *itemView `json:"item"`
}

func parseCursorPagination(c *gin.Context) (*cursorPosition, int, error) {
	pageSize := queryInt(c, "page_size", 10)
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	raw := strings.TrimSpace(c.Query("cursor"))
	if raw == "" {
		return nil, pageSize, nil
	}

	payload, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, pageSize, err
	}
	var cursor cursorPosition
	if err := json.Unmarshal(payload, &cursor); err != nil || cursor.ID == 0 || cursor.CreatedAt.IsZero() {
		return nil, pageSize, gorm.ErrInvalidData
	}
	return &cursor, pageSize, nil
}

func applyCursor(query *gorm.DB, cursor *cursorPosition) *gorm.DB {
	return applyCursorByID(query, cursor, "id")
}

func applyCursorByID(query *gorm.DB, cursor *cursorPosition, idColumn string) *gorm.DB {
	if cursor == nil {
		return query
	}
	return query.Where(
		"(created_at < ?) OR (created_at = ? AND "+idColumn+" < ?)",
		cursor.CreatedAt,
		cursor.CreatedAt,
		cursor.ID,
	)
}

func encodeCursor(createdAt time.Time, id uint) string {
	payload, _ := json.Marshal(cursorPosition{
		CreatedAt: createdAt,
		ID:        id,
	})
	return base64.RawURLEncoding.EncodeToString(payload)
}

func makePageResult[T any](
	list []T,
	total int64,
	pageSize int,
	rowCursor func(T) (time.Time, uint),
) pageResult {
	hasMore := len(list) > pageSize
	var nextCursor string
	if hasMore {
		list = list[:pageSize]
		last := list[len(list)-1]
		createdAt, id := rowCursor(last)
		nextCursor = encodeCursor(createdAt, id)
	}
	return pageResultWithCursor(list, total, pageSize, hasMore, nextCursor)
}

func pageResultWithCursor(
	list interface{},
	total int64,
	pageSize int,
	hasMore bool,
	nextCursor string,
) pageResult {
	return pageResult{
		List:       list,
		Total:      total,
		PageSize:   pageSize,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}
}

func queryInt(c *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return fallback
	}
	return value
}

func parseUint(value string) (uint64, error) {
	return strconv.ParseUint(value, 10, 64)
}

func parseTimeValue(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Now(), nil
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, gorm.ErrInvalidData
}

func validateItemInput(itemType, title, description, location, contact string, images []string) bool {
	if itemType != model.ItemTypeLost && itemType != model.ItemTypeFound {
		return false
	}
	if strings.TrimSpace(title) == "" || len([]rune(title)) > 50 {
		return false
	}
	if strings.TrimSpace(description) == "" || len([]rune(description)) > 5000 {
		return false
	}
	if strings.TrimSpace(location) == "" || len([]rune(location)) > 100 {
		return false
	}
	if strings.TrimSpace(contact) == "" || len([]rune(contact)) > 100 {
		return false
	}
	if len(images) > 6 {
		return false
	}
	for _, image := range images {
		if !strings.HasPrefix(image, "/uploads/") {
			return false
		}
	}
	return true
}

func itemToView(item model.Item) itemView {
	images := item.Images
	if images == nil {
		images = []string{}
	}
	view := itemView{
		ID:          item.ID,
		Type:        item.Type,
		Title:       item.Title,
		Description: item.Description,
		Location:    item.Location,
		LostAt:      item.LostAt,
		Contact:     item.Contact,
		Images:      images,
		Status:      item.Status,
		UID:         item.UID,
		Remark:      item.Remark,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
	if item.User != nil {
		view.User = userBrief{ID: item.User.UID, Username: item.User.Username}
	}
	return view
}

func claimToView(claim model.Claim) claimView {
	view := claimView{
		ID:        claim.ID,
		ItemID:    claim.ItemID,
		UID:       claim.UID,
		Proof:     claim.Proof,
		Contact:   claim.Contact,
		Status:    claim.Status,
		Remark:    claim.Remark,
		CreatedAt: claim.CreatedAt,
	}
	if claim.User != nil {
		view.Username = claim.User.Username
	}
	if claim.Item != nil {
		item := itemToView(*claim.Item)
		view.Item = &item
	}
	return view
}

func userToView(user model.User) gin.H {
	return gin.H{
		"uid":        user.UID,
		"username":   user.Username,
		"contact":    user.Contact,
		"role":       user.Role,
		"status":     user.Status,
		"created_at": user.CreatedAt,
	}
}

func respondDBError(c *gin.Context, err error) {
	if err == gorm.ErrRecordNotFound {
		response.Error(c, http.StatusNotFound, response.CodeNotFound, "资源不存在")
		return
	}
	response.Error(c, http.StatusInternalServerError, response.CodeServerError, "服务器内部错误")
}
