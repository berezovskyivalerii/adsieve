package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/oauth2"

	mw "github.com/berezovskyivalerii/adsieve/internal/delivery/rest/middleware"
	"github.com/berezovskyivalerii/adsieve/internal/domain/ports"
)

type Handler struct {
	userSvc    ports.User
	clickSvc   ports.Click
	convSvc    ports.Conversion
	metricsSvc ports.Metrics
	adsSvc     ports.Ads

	// интеграции Google
	oauthStates ports.OAuthStateStore
	tokenVault  ports.TokenVault
	gadsClient  ports.GoogleAdsClient
	oauthCfg    *oauth2.Config
}

func NewHandler(
	userSvc ports.User,
	clickSvc ports.Click,
	convSvc ports.Conversion,
	metricsSvc ports.Metrics,
	adsSvc ports.Ads,
	oauthStates ports.OAuthStateStore,
	tokenVault ports.TokenVault,
	gadsClient ports.GoogleAdsClient,
	oauthCfg *oauth2.Config,
) *Handler {
	return &Handler{
		userSvc:    userSvc,
		clickSvc:   clickSvc,
		convSvc:    convSvc,
		metricsSvc: metricsSvc,
		adsSvc:     adsSvc,

		oauthStates: oauthStates,
		tokenVault:  tokenVault,
		gadsClient:  gadsClient,
		oauthCfg:    oauthCfg,
	}
}

func (h *Handler) Router(jwtSecret []byte) http.Handler {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	jwtAuth := mw.NewJWTAuth(jwtSecret)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json")))
	r.GET("/api/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/api/swagger/doc.json")))

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

	r.POST("/integrations/google/connect", jwtAuth.Middleware(), h.googleConnect)
	// public
	r.GET("/integrations/google/callback", h.googleCallback)

	// private
	r.GET("/integrations/google/accounts", jwtAuth.Middleware(), h.googleAccounts)
	r.POST("/integrations/google/link-accounts", jwtAuth.Middleware(), h.googleLinkAccounts)

	return r
}
