package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
)

type signReq struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshReq struct {
	Refresh string `json:"refresh_token" binding:"required"`
}

// @Summary     Регистрация пользователя
// @Description Создаёт нового рекламодателя и сразу возвращает пару токенов (access + refresh).
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       input  body   signReq  true  "Данные регистрации (email, password)"
// @Success     201    {object}  map[string]string  "access_token, refresh_token"
// @Failure     400    {object}  map[string]string  "bad request / валидация входных данных"
// @Failure     409    {object}  map[string]string  "email_already_registered"
// @Failure     500    {object}  map[string]string  "internal error"
// @Router      /auth/sign-up [post]
func (h *Handler) signUp(c *gin.Context) {
	var user signReq
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := h.userSvc.SignUp(c.Request.Context(),
		entity.SignInput{
			Email:    user.Email,
			Password: user.Password,
		})

	switch err {
	case nil:
		c.JSON(http.StatusCreated, gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
	case errs.ErrEmailTaken:
		c.JSON(http.StatusConflict, gin.H{"error": "email_already_registered"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// @Summary     Логин пользователя
// @Description Аутентифицирует по email и паролю. Возвращает access и refresh токены.
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       input  body   signReq  true  "Данные входа (email, password)"
// @Success     200    {object}  map[string]string  "access_token, refresh_token"
// @Failure     400    {object}  map[string]string  "bad request / валидация входных данных"
// @Failure     401    {object}  map[string]string  "unauthorized (неверные учетные данные)"
// @Router      /auth/sign-in [post]
func (h *Handler) signIn(c *gin.Context) {
	var req signReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	acc, ref, err := h.userSvc.SignIn(
		c.Request.Context(),
		entity.SignInput{Email: req.Email, Password: req.Password},
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  acc,
		"refresh_token": ref,
	})
}

// @Summary     Обновление access-токена по refresh-токену
// @Description Принимает refresh_token и выдает новую пару токенов (access + refresh).
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       input  body   refreshReq  true  "Тело запроса с refresh_token"
// @Success     200    {object}  map[string]string  "access_token, refresh_token"
// @Failure     400    {object}  map[string]string  "bad request / валидация входных данных"
// @Failure     401    {object}  map[string]string  "unauthorized (refresh токен недействителен/просрочен)"
// @Router      /auth/refresh [post]
func (h *Handler) refresh(c *gin.Context) {
	var req refreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	acc, ref, err := h.userSvc.Refresh(c.Request.Context(), req.Refresh)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  acc,
		"refresh_token": ref,
	})
}
