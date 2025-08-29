// internal/delivery/rest/router.go
package rest

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	mw "github.com/berezovskyivalerii/adsieve/internal/delivery/rest/middleware"
	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
)

/* ===== сервисные интерфейсы ===== */

type User interface {
	SignUp(ctx context.Context, inp entity.SignInput) (string, string, error)
	SignIn(ctx context.Context, in entity.SignInput) (string, string, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
}

type Click interface {
	Click(ctx context.Context, clk entity.ClickInput) (int64, error)
}

type Conversion interface {
	Create(ctx context.Context, in entity.ConversionInput) (int64, error)
}

type Metrics interface {
	Get(ctx context.Context, userID int64, f entity.MetricsFilter) ([]entity.DailyMetricDTO, error)
}

type Ads interface {
	List(ctx context.Context, userID int64, f entity.AdsFilter) (items []entity.AdDTO, total int, err error)
}

type Handler struct {
	userSvc    User
	clickSvc   Click
	convSvc    Conversion
	metricsSvc Metrics
	adsSvc     Ads
}

func NewHandler(userSvc User, clickSvc Click, convSvc Conversion, metricsSvc Metrics, adsSvc Ads) *Handler {
	return &Handler{
		userSvc:    userSvc,
		clickSvc:   clickSvc,
		convSvc:    convSvc,
		metricsSvc: metricsSvc,
		adsSvc:     adsSvc,
	}
}

func (h *Handler) Router(jwtSecret []byte) http.Handler {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	jwtAuth := mw.NewJWTAuth(jwtSecret)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/sign-up", h.signUp)
			auth.POST("/sign-in", h.signIn)
			auth.POST("/refresh", h.refresh)
		}
		api.POST("/click", h.click)

		private := api.Group("/")
		private.Use(jwtAuth.Middleware())
		{
			private.POST("/conversion", h.conversion)
			private.GET("/metrics", h.metrics)
			private.GET("/ads", h.ads)
		}
	}
	return r
}
