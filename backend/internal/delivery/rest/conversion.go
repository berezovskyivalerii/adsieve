package rest

import (
	"time"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type convReq struct {
	ClickID    string          `json:"click_id"    binding:"required"`
	OrderValue decimal.Decimal `json:"order_value" binding:"required,dec_gt0"`
	OccurredAt int64           `json:"occurred_at"`
}

func (h *Handler) conversion(c *gin.Context) {
	var req convReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	id, err := h.convSvc.Create(c,
		entity.OrderInput{
			ClickID:    req.ClickID,
			OrderValue: req.OrderValue,
			OccurredAt: time.Unix(req.OccurredAt, 0).UTC(),
		})
	switch err {
	case nil:
		c.JSON(201, gin.H{"id": id})
	case errs.ErrClickNotFound:
		c.JSON(404, gin.H{"error": "click not found"})
	case errs.ErrOrderExists:
		c.JSON(409, gin.H{"error": "order already registered"})
	default:
		c.JSON(500, gin.H{"error": err.Error()})
	}
}
