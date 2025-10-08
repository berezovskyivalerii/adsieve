// cmd/main.go

// @title           AdSieve API
// @version         1.0
// @description     Backend система AdSieve (MVP). API для трекинга кликов, конверсий и аналитики.
// @termsOfService  http://swagger.io/terms/
// @contact.name   AdSieve Dev Team
// @contact.url    http://adsieve.example.com
// @contact.email  support@adsieve.com
// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT
// @host      localhost:8080
// @BasePath  /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	"golang.org/x/oauth2"

	"github.com/berezovskyivalerii/adsieve/internal/adapter/crypto"
	"github.com/berezovskyivalerii/adsieve/internal/adapter/googleads"
	"github.com/berezovskyivalerii/adsieve/internal/adapter/postgres"
	"github.com/berezovskyivalerii/adsieve/internal/delivery/rest"
	"github.com/berezovskyivalerii/adsieve/internal/domain/ports"
	"github.com/berezovskyivalerii/adsieve/internal/domain/service"
	"github.com/berezovskyivalerii/adsieve/internal/shared/googleoauth"

	_ "github.com/lib/pq"
)

type oauthCfgWrapper struct{ cfg *oauth2.Config }

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("dec_gt0", func(fl validator.FieldLevel) bool {
			d, ok := fl.Field().Interface().(decimal.Decimal)
			if !ok {
				return false
			}
			return d.GreaterThan(decimal.Zero)
		})
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf(".env not found: %v (ignored)", err)
	}

	// ============= 1) ENV =============
	dsn := mustEnv("DB_DSN")
	jwtSecret := []byte(mustEnv("JWT_SECRET"))
	httpPort := getenv("PORT", "8080")
	bcryptCost := atoi(getenv("BCRYPT_COST", ""))

	// Google OAuth env
	gcfg := googleoauth.Load()
	var oauthCfg *oauth2.Config
	if gcfg.ClientID != "" && gcfg.ClientSecret != "" && gcfg.RedirectURL != "" {
		oauthCfg = googleoauth.OAuth2(gcfg)
		log.Printf("Google OAuth configured: redirect=%s", gcfg.RedirectURL)
	} else {
		log.Printf("Google OAuth skipped: GOOGLE_CLIENT_ID/SECRET/REDIRECT_URL not set")
	}
	// Google Ads env
	devToken := getenv("GOOGLE_DEVELOPER_TOKEN", "")
	loginCID := getenv("GOOGLE_LOGIN_CUSTOMER_ID", "") // может быть пустым

	// ============= 2) DB =============
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("DB open: %v", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err = db.Ping(); err != nil {
		log.Fatalf("DB ping: %v", err)
	}

	// ============= 3) WIRING =============
	// репозитории
	userAdsRepo := postgres.NewUserAdsRepo(db)
	metricsRepo := postgres.NewMetricsRepo(db)
	convRepo := postgres.NewConversionRepo(db)
	clkRepo := postgres.NewClicksRepo(db)
	userRepo := postgres.NewUserRepo(db)
	tokenRepo := postgres.NewTokensRepo(db)
	adsRepo := postgres.NewAdsRepo(db)

	hasher := crypto.NewBcryptHasher(bcryptCost)
	stateRepo := postgres.NewOAuthStateRepo(db, 10*time.Minute)

	aead, err := crypto.NewAEADEncryptor(mustEnv("ENC_KEY"))
	if err != nil {
		log.Fatalf("encryptor: %v", err)
	}
	vaultRepo := postgres.NewTokenVault(db, aead)
	adAccRepo := postgres.NewGoogleAdAccountsRepo(db)

	// доменные сервисы
	metricsSvc := service.NewMetricsService(metricsRepo, userAdsRepo)
	convSvc := service.NewConversionService(clkRepo, convRepo)
	clkSvc := service.NewClickService(clkRepo)
	authSvc := service.NewAuthService(userRepo, tokenRepo, hasher, jwtSecret)
	adsSvc := service.NewAdsService(adsRepo)

	// === Google интеграции ===
	var (
		oauthStates ports.OAuthStateStore = stateRepo
		tokenVault  ports.TokenVault      = vaultRepo
		gadsClient  ports.GoogleAdsClient
		googleSync  rest.GoogleSync
	)

	if oauthCfg != nil && devToken != "" {
		ts := googleads.NewTokenSource(vaultRepo, oauthCfgWrapper{cfg: oauthCfg})
		gads := googleads.New(devToken, loginCID, ts)
		gadsClient = &gadsPortsAdapter{
			core:  gads,
			vault: vaultRepo,
			repo:  adAccRepo,
		}

		googleSync = service.NewGoogleSync(gads, adAccRepo)
	} else {
		log.Fatal("Google Ads not configured: set GOOGLE_CLIENT_ID/SECRET/REDIRECT_URL and GOOGLE_DEVELOPER_TOKEN")
	}

	// delivery
	handler := rest.NewHandler(
		authSvc,
		clkSvc,
		convSvc,
		metricsSvc,
		adsSvc,
		oauthStates,
		tokenVault,
		gadsClient,
		oauthCfg,
		googleSync,
	)

	// ============= 4) HTTP server =============
	srv := &http.Server{
		Addr:         ":" + httpPort,
		Handler:      handler.Router(jwtSecret),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("⇨ HTTP server started on :%s", httpPort)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server: %v", err)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⇨ Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("⇨ Server stopped")
}

type gadsPortsAdapter struct {
	core interface {
		ListAccessibleCustomers(ctx context.Context, userID int64) ([]string, string, error)
	}
	vault interface {
		LoadRefreshToken(ctx context.Context, userID int64) (googleUserID, refreshTokenEnc, scope string, err error)
	}
	repo interface {
		LinkGoogleAccounts(ctx context.Context, userID int64, tokenOwnerGoogleUserID string, customerIDs []string) error
	}
}

func (a *gadsPortsAdapter) ListAccessibleCustomers(ctx context.Context, userID int64) ([]string, error) {
	ids, _, err := a.core.ListAccessibleCustomers(ctx, userID) // игнорируем googleUID
	return ids, err
}

func (a *gadsPortsAdapter) LinkAccounts(ctx context.Context, userID int64, customerIDs []string) error {
	googleUID, _, _, err := a.vault.LoadRefreshToken(ctx, userID)
	if err != nil {
		return err
	}
	return a.repo.LinkGoogleAccounts(ctx, userID, googleUID, customerIDs)
}

func (w oauthCfgWrapper) ExchangeRefresh(ctx context.Context, refresh string) (*oauth2.Token, error) {
	t := &oauth2.Token{RefreshToken: refresh}
	return w.cfg.TokenSource(ctx, t).Token()
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("environment variable %s is required", key)
	}
	return val
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func atoi(s string) int {
	if s == "" {
		return 0
	}
	n, _ := strconv.Atoi(s)
	return n
}
