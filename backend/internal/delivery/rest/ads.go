// internal/delivery/rest/ads_handler.go
package rest

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
)

type AdsListMeta struct {
	Page     int  `json:"page"`
	PageSize int  `json:"page_size"`
	Total    int  `json:"total"`
	HasMore  bool `json:"has_more"`
}

type AdsListResponse struct {
	Items []entity.AdDTO `json:"items"`
	Meta  AdsListMeta    `json:"meta"`
}

// @Summary     Список объявлений пользователя
// @Description Возвращает объявления текущего пользователя с фильтрами, пагинацией и сортировкой.
// @Tags        Ads
// @Produce     json
// @Security    BearerAuth
// @Param       status     query   string  false  "Фильтр по статусу"        Enums(active,paused,all) default(all)
// @Param       platform   query   string  false  "Платформа"                Enums(facebook,google)
// @Param       q          query   string  false  "Подстрока для поиска по названию"
// @Param       ad_id      query   []int64 false  "Список ad_id (CSV)"       collectionFormat(csv)
// @Param       page       query   int     false  "Номер страницы (>=1)"     default(1)
// @Param       page_size  query   int     false  "Размер страницы (1..200)" default(50)
// @Param       sort       query   string  false  "Сортировка"               Enums(name,-name,created_at,-created_at) default(name)
// @Success     200        {object}  AdsListResponse
// @Failure     400        {object}  map[string]string  "invalid status | invalid ad_id list"
// @Failure     401        {object}  map[string]string  "unauthorized"
// @Failure     500        {object}  map[string]string  "internal error"
// @Router      /ads [get]
func (h *Handler) ads(c *gin.Context) {
	userID, ok := getUserID(c) // ← единый способ извлечь ID из контекста
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

	resp := AdsListResponse{
		Items: items,
		Meta: AdsListMeta{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
			HasMore:  f.Offset+len(items) < total,
		},
	}
	c.JSON(http.StatusOK, resp)
}

/* ===== helpers ===== */

const ctxUserKey = "userID"

// getUserID достаёт userID из контекста, поддерживая оба ключа: "userID" и "user_id".
func getUserID(c *gin.Context) (int64, bool) {
	if id, ok := extractID(c, ctxUserKey); ok {
		return id, true
	}
	if id, ok := extractID(c, "user_id"); ok {
		return id, true
	}
	return 0, false
}

func extractID(c *gin.Context, key string) (int64, bool) {
	v, ok := c.Get(key)
	if !ok {
		return 0, false
	}
	switch t := v.(type) {
	case int64:
		return t, true
	case int:
		return int64(t), true
	case float64:
		return int64(t), true
	case string:
		id, err := strconv.ParseInt(t, 10, 64)
		return id, err == nil
	default:
		return 0, false
	}
}

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
