package rest

import (
	"context"
	"net/http"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	"github.com/gin-gonic/gin"
)

type User interface {
	SignUp(ctx context.Context, inp entity.SignInput) (string, string, error)
	SignIn(ctx context.Context, in entity.SignInput) (string, string, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
}

type Handler struct {
	userSvc User
}

func NewHandler(userSvc User) *Handler {
	return &Handler{userSvc: userSvc}
}

func (h *Handler) Router() http.Handler {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	auth := r.Group("/api/auth")
	{
		auth.POST("/sign-up", h.signUp)
		auth.POST("/sign-in", h.signIn)
		auth.POST("/refresh", h.refresh)
	}

	return r
}
