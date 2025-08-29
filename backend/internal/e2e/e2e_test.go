package e2e_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

const (
	defaultBaseURL = "http://localhost:8080"
	testAdID = int64(101)
)

// --------- DTOs ---------

type signInResp struct {
	AccessToken string `json:"access_token"`
}

type apiError struct {
	Error string `json:"error"`
}

type metricItem struct {
	AdID        int64  `json:"ad_id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Day         string `json:"day"`
	Clicks      int    `json:"clicks"`
	Conversions int    `json:"conversions"`
	RevenueStr  string `json:"revenue"`
	SpendStr    string `json:"spend"`
	RoasStr     string `json:"roas"`
}

// --------- Helpers ---------

func mustEnv(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func mustReadFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read file %s: %v", p, err)
	}
	return string(b)
}

func httpJSON(t *testing.T, method, url string, body string, headers map[string]string) (int, []byte) {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("http do %s %s: %v", method, url, err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return res.StatusCode, b
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// --------- DB helpers ---------

func openDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		t.Skip("DB_DSN is empty — skip E2E")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("db.Ping: %v", err)
	}
	return db
}

func ensureUserAds(t *testing.T, db *sql.DB, email string, adID int64) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var uid int64
	if err := db.QueryRowContext(ctx, `SELECT user_id FROM users WHERE email=$1`, email).Scan(&uid); err != nil {
		t.Fatalf("get user_id: %v", err)
	}

	_, err := db.ExecContext(ctx, `INSERT INTO user_ads (user_id, ad_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, uid, adID)
	if err != nil {
		t.Fatalf("insert user_ads: %v", err)
	}
}

func runAggregatorSQL(t *testing.T, db *sql.DB) {
	t.Helper()
	// Запускаем ровно тот же SQL, что и продовый агрегатор.
	// Путь относительно корня модуля:
	sqlPath := filepath.Join("sql", "aggregate_daily.sql")
	sqlText := mustReadFile(t, sqlPath)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := db.ExecContext(ctx, sqlText); err != nil {
		t.Fatalf("aggregate sql exec: %v", err)
	}
}

// --------- The Test ---------

func TestE2E_Click_Conversion_Aggregate_Metrics(t *testing.T) {
	baseURL := mustEnv("BASE_URL", defaultBaseURL)
	db := openDB(t)
	defer db.Close()

	// 1) Создаём тестового пользователя (уникальный email) и логинимся
	email := fmt.Sprintf("e2e_%s@adsieve.local", randomHex(5))
	password := "test123"

	// sign-up (идемпотентно)
	_, _ = httpJSON(t, http.MethodPost, baseURL+"/api/auth/sign-up",
		fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password), nil)

	// sign-in
	_, body := httpJSON(t, http.MethodPost, baseURL+"/api/auth/sign-in",
		fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password), nil)

	var si signInResp
	if err := json.Unmarshal(body, &si); err != nil || si.AccessToken == "" {
		t.Fatalf("sign-in failed, body=%s, err=%v", string(body), err)
	}
	authz := map[string]string{"Authorization": "Bearer " + si.AccessToken}

	// 2) Связываем пользователя с нужным объявлением (authorization в /metrics)
	ensureUserAds(t, db, email, testAdID)

	// 3) Отправляем клик
	clickID := "click_" + strconv.FormatInt(time.Now().Unix(), 10)
	clickAtSec := time.Now().UTC().Unix() // seconds
	// основной путь — seconds
	status, body := httpJSON(t, http.MethodPost, baseURL+"/api/click",
		fmt.Sprintf(`{"click_id":"%s","ad_id":%d,"clicked_at":%d}`, clickID, testAdID, clickAtSec),
		map[string]string{"Content-Type": "application/json"},
	)
	if status >= 400 {
		// fallback: попробуем миллисекунды, если когда-то контракт поменяют
		clickAtMS := time.Now().UTC().UnixMilli()
		status2, body2 := httpJSON(t, http.MethodPost, baseURL+"/api/click",
			fmt.Sprintf(`{"click_id":"%s","ad_id":%d,"clicked_at":%d}`, clickID, testAdID, clickAtMS),
			map[string]string{"Content-Type": "application/json"},
		)
		if status2 >= 400 {
			t.Fatalf("click failed. status=%d body=%s; fallback status=%d body=%s", status, string(body), status2, string(body2))
		}
	}

	// 4) Отправляем конверсию (идемпотентно по order_id)
	orderID := "ORD-" + randomHex(4)
	status, body = httpJSON(t, http.MethodPost, baseURL+"/api/conversion",
		fmt.Sprintf(`{"order_id":"%s","revenue":49.99,"click_id":"%s"}`, orderID, clickID),
		authz,
	)
	if status >= 400 {
		var ae apiError
		_ = json.Unmarshal(body, &ae)
		// допустим 409 (или свой текст ошибки на idempotency)
		if status != http.StatusConflict && !strings.Contains(strings.ToLower(ae.Error), "exists") {
			t.Fatalf("conversion create failed: status=%d body=%s", status, string(body))
		}
	}

	// Повторим ту же конверсию (проверка идемпотентности; ожидаем 200/409/эквивалент)
	status, _ = httpJSON(t, http.MethodPost, baseURL+"/api/conversion",
		fmt.Sprintf(`{"order_id":"%s","revenue":49.99,"click_id":"%s"}`, orderID, clickID),
		authz,
	)
	if !(status == http.StatusOK || status == http.StatusConflict || status == http.StatusCreated) {
		t.Fatalf("conversion idempotency unexpected status=%d", status)
	}

	// 5) Запускаем агрегаторный SQL
	runAggregatorSQL(t, db)

	// 6) Читаем метрики за сегодня
	today := time.Now().UTC().Format("2006-01-02")
	_, body = httpJSON(t, http.MethodGet, fmt.Sprintf("%s/api/metrics?ad_ids=%d&from=%s&to=%s", baseURL, testAdID, today, today), "", authz)

	var items []metricItem
	if err := json.Unmarshal(body, &items); err != nil {
		t.Fatalf("metrics json parse: %v; body=%s", err, string(body))
	}
	if len(items) == 0 {
		t.Fatalf("metrics empty for ad_id=%d on %s", testAdID, today)
	}
	var got *metricItem
	for i := range items {
		if items[i].AdID == testAdID && items[i].Day == today {
			got = &items[i]
			break
		}
	}
	if got == nil {
		t.Fatalf("metrics row not found for ad_id=%d day=%s; items=%v", testAdID, today, items)
	}

	// Проверки значений
	if got.Clicks < 1 {
		t.Fatalf("expected clicks>=1, got=%d", got.Clicks)
	}
	if got.Conversions < 1 {
		t.Fatalf("expected conversions>=1, got=%d", got.Conversions)
	}
	if got.RevenueStr != "49.99" {
		// допускаем форматирование с нулями (49.9900)
		if !strings.HasPrefix(got.RevenueStr, "49.99") {
			t.Fatalf("expected revenue=49.99, got=%s", got.RevenueStr)
		}
	}

	// 7) Повторный агрегатор не должен менять показатели (идемпотентно)
	before := *got
	runAggregatorSQL(t, db)

	_, body = httpJSON(t, http.MethodGet, fmt.Sprintf("%s/api/metrics?ad_ids=%d&from=%s&to=%s", baseURL, testAdID, today, today), "", authz)
	_ = json.Unmarshal(body, &items)
	var after *metricItem
	for i := range items {
		if items[i].AdID == testAdID && items[i].Day == today {
			after = &items[i]
			break
		}
	}
	if after == nil {
		t.Fatalf("metrics row disappeared after re-aggregate")
	}
	if before.Clicks != after.Clicks || before.Conversions != after.Conversions || before.RevenueStr != after.RevenueStr {
		t.Fatalf("idempotency broken: before=%+v after=%+v", before, after)
	}
}

// небольшой smoke, чтобы быстро подсказать, почему BASE_URL не отвечает
func TestAPIAlive(t *testing.T) {
	baseURL := mustEnv("BASE_URL", defaultBaseURL)
	code, _ := httpJSON(t, http.MethodGet, baseURL+"/api/click", "", nil)
	if !(code == http.StatusNotFound || code == http.StatusBadRequest || code == http.StatusMethodNotAllowed) {
		t.Fatalf("api not responding as expected, status=%d", code)
	}
}

// уточняющий smoke для БД
func TestDBAlive(t *testing.T) {
	db := openDB(t)
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var one int
	if err := db.QueryRowContext(ctx, `SELECT 1`).Scan(&one); err != nil || one != 1 {
		t.Fatalf("db basic query failed: %v, one=%d", err, one)
	}
}

// маленький хелпер для ошибок API (когда body={"error":...})
// func parseAPIError(b []byte) (string, bool) {
// 	var e apiError
// 	if err := json.Unmarshal(b, &e); err == nil && e.Error != "" {
// 		return e.Error, true
// 	}
// 	return "", false
// }

// func requireStatusOK(t *testing.T, status int, body []byte) {
// 	t.Helper()
// 	if status >= 400 {
// 		if msg, ok := parseAPIError(body); ok {
// 			t.Fatalf("unexpected api error: %s", msg)
// 		}
// 		t.Fatalf("unexpected status=%d body=%s", status, string(body))
// 	}
// }

// func ignoreErr(err error) {
// 	_ = err
// }

// func isErrNoRows(err error) bool {
// 	return errors.Is(err, sql.ErrNoRows)
// }
