package rest

import (
	"net/http"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
	"github.com/gin-gonic/gin"
)

type signReq struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshReq struct {
	Refresh string `json:"refresh_token" binding:"required"`
}

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
