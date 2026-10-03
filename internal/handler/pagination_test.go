package handler

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestCursorRoundTrip(t *testing.T) {
	createdAt := time.Date(2026, 9, 28, 12, 34, 56, 123000000, time.FixedZone("CST", 8*60*60))
	raw := encodeCursor(createdAt, 42)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/items?page_size=20&cursor="+raw, nil)
	cursor, pageSize, err := parseCursorPagination(c)
	if err != nil {
		t.Fatalf("parseCursorPagination() error = %v", err)
	}
	if pageSize != 20 {
		t.Fatalf("page size = %d, want 20", pageSize)
	}
	if cursor.ID != 42 || !cursor.CreatedAt.Equal(createdAt) {
		t.Fatalf("cursor = %+v, want id 42 and time %v", cursor, createdAt)
	}
}

func TestMakePageResultTrimsLookaheadRow(t *testing.T) {
	type row struct {
		ID        uint
		CreatedAt time.Time
	}
	rows := []row{
		{ID: 3, CreatedAt: time.Date(2026, 9, 28, 12, 0, 2, 0, time.UTC)},
		{ID: 2, CreatedAt: time.Date(2026, 9, 28, 12, 0, 1, 0, time.UTC)},
		{ID: 1, CreatedAt: time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)},
	}

	result := makePageResult(rows, 8, 2, func(item row) (time.Time, uint) {
		return item.CreatedAt, item.ID
	})
	if len(result.List.([]row)) != 2 {
		t.Fatalf("result list length = %d, want 2", len(result.List.([]row)))
	}
	if !result.HasMore || result.NextCursor == "" {
		t.Fatalf("result = %+v, want next cursor", result)
	}
	if result.Total != 8 {
		t.Fatalf("result total = %d, want 8", result.Total)
	}
}
