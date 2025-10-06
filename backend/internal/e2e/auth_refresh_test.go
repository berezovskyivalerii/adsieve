//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type authResp struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
}

func baseURL() string {
	if v := os.Getenv("BASE_URL"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return "http://localhost:8080"
}

func postJSON(t *testing.T, path string, body any) (*http.Response, []byte) {
	t.Helper()
	var rd io.Reader
	switch b := body.(type) {
	case []byte:
		rd = bytes.NewReader(b)
	case string:
		rd = strings.NewReader(b)
	default:
		data, err := json.Marshal(body)
		require.NoError(t, err)
		rd = bytes.NewReader(data)
	}
	u, _ := url.JoinPath(baseURL(), path)
	req, err := http.NewRequest(http.MethodPost, u, rd)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	cli := &http.Client{Timeout: 7 * time.Second}
	res, err := cli.Do(req)
	require.NoError(t, err)
	all, _ := io.ReadAll(res.Body)
	res.Body.Close()
	return res, all
}

func getAuth(t *testing.T, path string, access string) (*http.Response, []byte) {
	t.Helper()
	u, _ := url.JoinPath(baseURL(), path)
	req, _ := http.NewRequest(http.MethodGet, u, nil)
	if access != "" {
		req.Header.Set("Authorization", "Bearer "+access)
	}
	cli := &http.Client{Timeout: 7 * time.Second}
	res, err := cli.Do(req)
	require.NoError(t, err)
	all, _ := io.ReadAll(res.Body)
	res.Body.Close()
	return res, all
}

func ensureTokens(t *testing.T, email, pass string) authResp {
	// попытка sign-up
	res, body := postJSON(t, "/api/auth/sign-up", map[string]string{
		"email":    email,
		"password": pass,
	})
	switch res.StatusCode {
	case http.StatusCreated:
		var ar authResp
		require.NoError(t, json.Unmarshal(body, &ar), "sign-up json: %s", string(body))
		require.NotEmpty(t, ar.Access)
		require.NotEmpty(t, ar.Refresh)
		return ar
	case http.StatusConflict:
		// уже есть — делаем sign-in
		res, body = postJSON(t, "/api/auth/sign-in", map[string]string{
			"email":    email,
			"password": pass,
		})
		require.Equal(t, http.StatusOK, res.StatusCode, "sign-in failed: %s", string(body))
		var ar authResp
		require.NoError(t, json.Unmarshal(body, &ar))
		require.NotEmpty(t, ar.Access)
		require.NotEmpty(t, ar.Refresh)
		return ar
	default:
		t.Fatalf("unexpected sign-up status: %d body=%s", res.StatusCode, string(body))
		return authResp{}
	}
}

func TestAuth_Refresh_Strict(t *testing.T) {
	email := fmt.Sprintf("e2e+%d@adsieve.local", time.Now().UnixNano())
	pass := "superpass"

	// 0) sanity: приватка без токена = 401
	res, _ := getAuth(t, "/api/ads", "")
	require.Equal(t, http.StatusUnauthorized, res.StatusCode, "private must require JWT")

	// 1) заведём пользователя и получим пару токенов
	ar := ensureTokens(t, email, pass)

	// приватка с access должна отвечать 200
	res, body := getAuth(t, "/api/ads", ar.Access)
	require.Equal(t, http.StatusOK, res.StatusCode, "private with access must be 200, body=%s", string(body))

	// 2) негативные кейсы refresh

	// 2.1) пустое тело
	res, _ = postJSON(t, "/api/auth/refresh", "")
	require.Equal(t, http.StatusBadRequest, res.StatusCode, "empty body must be 400")

	// 2.2) битый JSON
	res, _ = postJSON(t, "/api/auth/refresh", "{bad json")
	require.Equal(t, http.StatusBadRequest, res.StatusCode, "invalid json must be 400")

	// 2.3) случайная строка вместо refresh
	res, _ = postJSON(t, "/api/auth/refresh", map[string]string{"refresh": "not-a-token"})
	require.Equal(t, http.StatusUnauthorized, res.StatusCode, "random string refresh must be 401")

	// 2.4) подсовываем ACCESS как refresh → 401
	res, _ = postJSON(t, "/api/auth/refresh", map[string]string{"refresh": ar.Access})
	require.Equal(t, http.StatusUnauthorized, res.StatusCode, "access used as refresh must be 401")

	// 3) валидный refresh → новая пара (ротация)
	res, body = postJSON(t, "/api/auth/refresh", map[string]string{"refresh": ar.Refresh})
	require.Equal(t, http.StatusOK, res.StatusCode, "refresh failed: %s", string(body))

	var ar2 authResp
	require.NoError(t, json.Unmarshal(body, &ar2))
	require.NotEmpty(t, ar2.Access)
	require.NotEmpty(t, ar2.Refresh)
	require.NotEqual(t, ar.Access, ar2.Access, "access must be rotated")
	require.NotEqual(t, ar.Refresh, ar2.Refresh, "refresh must be rotated")

	// приватка с НОВЫМ access = 200
	res, body = getAuth(t, "/api/ads", ar2.Access)
	require.Equal(t, http.StatusOK, res.StatusCode, "new access must work, body=%s", string(body))

	// 4) повторное использование СТАРОГО refresh должно быть запрещено
	res, _ = postJSON(t, "/api/auth/refresh", map[string]string{"refresh": ar.Refresh})
	require.Equal(t, http.StatusUnauthorized, res.StatusCode, "old refresh must be invalid after rotation")

	// 5) повторная ротация по новому refresh (цепочка должна работать)
	res, body = postJSON(t, "/api/auth/refresh", map[string]string{"refresh": ar2.Refresh})
	require.Equal(t, http.StatusOK, res.StatusCode, "second rotation failed: %s", string(body))
	var ar3 authResp
	require.NoError(t, json.Unmarshal(body, &ar3))
	require.NotEmpty(t, ar3.Access)
	require.NotEmpty(t, ar3.Refresh)
	require.NotEqual(t, ar2.Access, ar3.Access)
	require.NotEqual(t, ar2.Refresh, ar3.Refresh)

	// приватка с третьим access = 200
	res, _ = getAuth(t, "/api/ads", ar3.Access)
	require.Equal(t, http.StatusOK, res.StatusCode, "third access must work")
}
