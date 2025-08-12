package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"database/sql"

	"github.com/berezovskyivalerii/adsieve/internal/adapter/crypto"
	"github.com/berezovskyivalerii/adsieve/internal/adapter/postgres"
	"github.com/berezovskyivalerii/adsieve/internal/delivery/rest"
	"github.com/berezovskyivalerii/adsieve/internal/domain/service"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

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

	// 1. Чтение переменных окружения
	dsn := mustEnv("DB_DSN")
	jwtSecret := []byte(mustEnv("JWT_SECRET"))
	httpPort := getenv("PORT", "8080")
	bcryptCost := atoi(getenv("BCRYPT_COST", ""))

	// 2. Подключение к базе
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

	// 3. Сборка зависимостей
	userAdsRepo := postgres.NewUserAdsRepo(db)
	metricsRepo := postgres.NewMetricsRepo(db)
	convRepo := postgres.NewConversionRepo(db)
	clkRepo := postgres.NewClicksRepo(db)
	userRepo := postgres.NewUserRepo(db)
	tokenRepo := postgres.NewTokensRepo(db)
	hasher := crypto.NewBcryptHasher(bcryptCost)

	metricsSvc := service.NewMetricsService(metricsRepo, userAdsRepo)
	convSvc := service.NewConversionService(clkRepo, convRepo)
	clkSvc := service.NewClickService(clkRepo)
	authSvc := service.NewAuthService(userRepo, tokenRepo, hasher, jwtSecret)
	
	handler := rest.NewHandler(authSvc, clkSvc, convSvc, metricsSvc)

	// 4. HTTP-сервер + graceful shutdown
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

	// ждём SIGINT/SIGTERM
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
