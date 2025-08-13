package rest

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ads(c *gin.Context){
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

	platform := strings.ToLower(strings.TrimSpace(c.Query("platform")))
	q := strings.TrimSpace(c.Query("q"))
	adIDs, err := parseCSVInt64(c.Query("ad_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ad_id list"})
		return
	}
}