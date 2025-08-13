package rest

import (
	"net/http"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
	"github.com/gin-gonic/gin"
)

type clickReq struct {
	ClickID   string `json:"click_id" binding:"required"`
	AdID      int64  `json:"ad_id" binding:"required"`
	ClickedAt int64  `json:"clicked_at" binding:"required"`
}

// Регистрация клика, сохранение в БД
func (h *Handler) click(c *gin.Context) {
	var req clickReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clk := entity.ClickInput{
		ClickID:   req.ClickID,
		AdID:      req.AdID,
		ClickedAt: &req.ClickedAt,
	}

	id, err := h.clickSvc.Click(c.Request.Context(), clk)
	switch err {
	case nil:
		c.JSON(http.StatusCreated, gin.H{"click_id": id})
	case errs.ErrDuplicateClick:
		c.JSON(http.StatusConflict, gin.H{"error": "click_already_registered"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
