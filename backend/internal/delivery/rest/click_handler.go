package rest

import (
	"net/http"
	"time"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
	"github.com/gin-gonic/gin"
)

type clickReq struct {
	ClickID    string `json:"click_id" binding:"required"`
	AdID       int64  `json:"ad_id" binding:"required"`
	OccurredAt int64  `json:"occurred_at" binding:"required"`
}

func (h *Handler) click(c *gin.Context) {
	var req clickReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clk := entity.Click{
		ClickID:    req.ClickID,
		AdID:       req.AdID,
		OccurredAt: time.Unix(req.OccurredAt, 0).UTC(),
	}

	id, err := h.clickSvc.Click(c.Request.Context(), clk)
	switch err {
	case nil:
		c.JSON(http.StatusCreated, gin.H{"click_id": id})
	case errs.ErrDuplicateClick:
		c.JSON(http.StatusConflict, gin.H{"error": "click already registered"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
