// internal/transport/rest/metrics_handler.go
package rest

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
	"github.com/gin-gonic/gin"
)

/*
GET /api/metrics?ad_id=87,91&from=2025-07-01&to=2025-07-19
Headers: Authorization: Bearer <JWT>

Поведение:
  - Если query-параметры отсутствуют, отдаём «с 1-го числа текущего месяца до вчера».
  - ad_id может быть пустым или списком через запятую.
*/
func (h *Handler) metrics(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// parse query
	from, to, err := parseDateRange(
		c.Query("from"),
		c.Query("to"),
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_date_range"})
		return
	}

	adIDs, err := parseAdIDs(c.Query("ad_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_ad_id"})
		return
	}

	filter := entity.MetricsFilter{
		AdIDs: adIDs,
		From:  from,
		To:    to,
	}

	// call service
	list, err := h.metricsSvc.Get(c.Request.Context(), userID.(int64), filter)
	switch err {
	case nil:
		c.JSON(http.StatusOK, list)
	case errs.ErrInvalidRange:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_date_range"})
	case errs.ErrNoAdAccess:
		c.JSON(http.StatusNotFound, gin.H{"error": "ad_not_found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// parseDateRange returns [from,to] or default values (current-month).
func parseDateRange(fromStr, toStr string) (time.Time, time.Time, error) {
	const template = "2006-01-02"
	now := time.Now().UTC()

	if fromStr == "" {
		fromStr = now.Format(template)[:8] + "01" // 1th month
	}
	if toStr == "" {
		toStr = now.AddDate(0, 0, -1).Format(template) // yesterday
	}
	from, err1 := time.ParseInLocation(template, fromStr, time.UTC)
	to, err2 := time.ParseInLocation(template, toStr, time.UTC)
	if err1 != nil || err2 != nil {
		return time.Time{}, time.Time{}, err1 // whatever
	}
	return from, to, nil
}

// parseAdIDs "87,91,92" → []int64{87,91,92}. Empty string -> nil.
func parseAdIDs(raw string) ([]int64, error) {
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	out := make([]int64, 0, len(parts))
	for _, p := range parts {
		id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if err != nil || id <= 0 {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}
