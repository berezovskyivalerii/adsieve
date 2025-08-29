// internal/delivery/rest/ads_handler.go
package rest

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
)

// GET /api/ads
// Query (все опц.): status=active|paused|all, platform=facebook|google, q=substr,
// ad_id=1,2,3, page=1, page_size=50, sort=name|-name|created_at|-created_at
func (h *Handler) ads(c *gin.Context) {
	uidVal, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, ok := uidVal.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	status := strings.ToLower(strings.TrimSpace(c.DefaultQuery("status", "all")))
	switch status {
	case "all", "active", "paused":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status (use: all|active|paused)"})
		return
	}
	platform := strings.ToLower(strings.TrimSpace(c.Query("platform"))) // "" | facebook | google
	q := strings.TrimSpace(c.Query("q"))
	adIDs, err := parseCSVInt64(c.Query("ad_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ad_id list"})
		return
	}

	page := mustAtoiDefault(c.Query("page"), 1)
	if page < 1 {
		page = 1
	}
	pageSize := mustAtoiDefault(c.Query("page_size"), 50)
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}
	sort := c.DefaultQuery("sort", "name")
	switch sort {
	case "name", "-name", "created_at", "-created_at":
	default:
		sort = "name"
	}

	f := entity.AdsFilter{
		Status:   statusIf(status != "all", status),
		Platform: stringPtrIf(platform != "", platform),
		Query:    stringPtrIf(q != "", q),
		AdIDs:    adIDs,
		Limit:    pageSize,
		Offset:   (page - 1) * pageSize,
		Sort:     sort,
	}

	items, total, err := h.adsSvc.List(c.Request.Context(), userID, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	hasMore := f.Offset+len(items) < total
	c.JSON(http.StatusOK, gin.H{
		"items": items, // []entity.AdDTO (ad_id, name, status, platform)
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
			"has_more":  hasMore,
		},
	})
}

/* ===== helpers ===== */

func parseCSVInt64(s string) ([]int64, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	out := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

func mustAtoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

func stringPtrIf(cond bool, v string) *string {
	if !cond {
		return nil
	}
	return &v
}
func statusIf(cond bool, v string) *string { return stringPtrIf(cond, v) }
