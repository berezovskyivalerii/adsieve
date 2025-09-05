package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
)

type clickReq struct {
	ClickID   string `json:"click_id" binding:"required"`
	AdID      int64  `json:"ad_id" binding:"required"`
	ClickedAt int64  `json:"clicked_at" binding:"required"`
}

// @Summary     Регистрация клика
// @Description Публичный эндпоинт. Принимает событие клика по объявлению и сохраняет в БД.
// @Description Поля: click_id (уникально), ad_id (ID объявления), clicked_at (UNIX UTC).
// @Tags        Tracking
// @Accept      json
// @Produce     json
// @Param       input  body   clickReq  true  "Данные клика"
// @Success     201    {object}  map[string]interface{}  "Пример: {\"click_id\": 12345}"
// @Failure     400    {object}  map[string]string       "bad request / валидация"
// @Failure     409    {object}  map[string]string       "click_already_registered"
// @Failure     500    {object}  map[string]string       "internal error"
// @Router      /click [post]
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
