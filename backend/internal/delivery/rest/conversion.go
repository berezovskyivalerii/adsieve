package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
)

type conversionRequest struct {
	ClickID     string  `json:"click_id"      binding:"required"`      // ID клика
	Revenue     float64 `json:"revenue"       binding:"required,gt=0"` // Прибыль
	OrderID     *string `json:"order_id,omitempty"`                    // ID заказа в системе магазина
	ConvertedAt *int64  `json:"converted_at,omitempty"`                // Время совершения заказа
}

// @Summary     Регистрация конверсии (заказа)
// @Description Публичный эндпоинт. Принимает конверсию, связанную с ранее зарегистрированным кликом (по click_id).
// @Description Поля:
// @Description - click_id (required) — ID клика, который уже существует в БД
// @Description - revenue  (required, >0) — сумма заказа/ценность конверсии
// @Description - order_id (optional) — ID заказа в магазине (для дедупликации)
// @Description - converted_at (optional) — UNIX-таймстамп, когда произошла конверсия
// @Tags        Tracking
// @Accept      json
// @Produce     json
// @Param       input  body   conversionRequest  true  "Данные конверсии"
// @Success     201    {object}  map[string]interface{}  "Пример: {\"conversion_id\": 12345}"
// @Failure     400    {object}  map[string]string       "bad request / валидация входных данных"
// @Failure     404    {object}  map[string]string       "click_no_found — клик по такому click_id не найден"
// @Failure     409    {object}  map[string]string       "conversion_already_registered — конверсия с таким order_id уже учтена"
// @Failure     500    {object}  map[string]string       "internal error"
// @Router      /conversion [post]
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
