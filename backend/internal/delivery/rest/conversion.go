package rest

import (
	"net/http"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type conversionRequest struct {
	ClickID     string  `json:"click_id"      binding:"required"` // ID клика
	Revenue     float64 `json:"revenue"       binding:"required,gt=0"` // Прибыль
	OrderID     *string `json:"order_id,omitempty"` // ID заказа в системе магазина
	ConvertedAt *int64  `json:"converted_at,omitempty"` // Время совершения заказа
}

// Регистрация конверсии (заказа), добавление в БД
func (h *Handler) conversion(c *gin.Context) {
	var req conversionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	in := entity.ConversionInput{
		ClickID:     req.ClickID,
		Revenue:     decimal.NewFromFloat(req.Revenue),
		OrderID:     req.OrderID,
		ConvertedAt: req.ConvertedAt,
	}

	conversionID, err := h.convSvc.Create(c, in)
	switch err {
	case nil:
		c.JSON(http.StatusCreated, gin.H{"conversion_id": conversionID})
	case errs.ErrClickNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "click_no_found"})
	case errs.ErrConversionExists:
		c.JSON(http.StatusConflict, gin.H{"error": "conversion_already_registered"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
