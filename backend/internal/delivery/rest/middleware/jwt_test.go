package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	mw "github.com/berezovskyivalerii/adsieve/internal/delivery/rest/middleware"
)

func TestJWTMiddleware_ValidToken(t *testing.T) {
	secret := []byte("test-secret")
	j := mw.NewJWTAuth(secret)

	// генерируем токен с sub=42 и exp через 1ч
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mw.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "42",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	raw, _ := token.SignedString(secret)

	r := gin.New()
	r.Use(j.Middleware())
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Authorization", "Bearer "+raw)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "pong", w.Body.String())
}
