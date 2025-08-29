package main

import (
	"context"
	"database/sql"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

const lockKey int64 = 1002 // ключ для pg_advisory_lock

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	dsn := getenv("DB_DSN", "postgres://user:pass@db:5432/adsieve?sslmode=disable")
	mockAds := getenv("MOCK_ADS", "")
	shiftDays, _ := strconv.Atoi(getenv("MOCK_DATE_SHIFT_DAYS", "0"))
	minSpend, _ := strconv.ParseFloat(getenv("MOCK_SPEND_MIN", "5"), 64)
	maxSpend, _ := strconv.ParseFloat(getenv("MOCK_SPEND_MAX", "50"), 64)

	db, err := openDB(dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := withLock(ctx, db, lockKey, func(ctx context.Context) error {
		adIDs, err := resolveAdIDs(ctx, db, mockAds)
		if err != nil {
			return err
		}
		if len(adIDs) == 0 {
			log.Printf("no ads to sync, exit")
			return nil
		}

		day := time.Now().AddDate(0, 0, shiftDays).UTC().Truncate(24 * time.Hour)

		for _, ad := range adIDs {
			spend := randomSpend(minSpend, maxSpend)
			_, err := db.ExecContext(ctx, `
				INSERT INTO ads_insights (ad_id, insight_date, spend)
				VALUES ($1, $2::date, $3)
				ON CONFLICT (ad_id, insight_date) DO UPDATE
				SET spend = EXCLUDED.spend
			`, ad, day, spend)
			if err != nil {
				return err
			}
			log.Printf("upsert insights: ad_id=%d date=%s spend=%.2f",
				ad, day.Format("2006-01-02"), spend)
		}
		return nil
	}); err != nil {
		log.Fatalf("sync_fb_insights failed: %v", err)
	}

	log.Printf("sync_fb_insights OK")
}

// --- helpers ---

func resolveAdIDs(ctx context.Context, db *sql.DB, mock string) ([]int64, error) {
	if strings.TrimSpace(mock) != "" {
		return parseIDs(mock), nil
	}
	rows, err := db.QueryContext(ctx, `SELECT ad_id FROM ads`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func parseIDs(s string) []int64 {
	parts := strings.Split(s, ",")
	out := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if v, err := strconv.ParseInt(p, 10, 64); err == nil && v > 0 {
			out = append(out, v)
		}
	}
	return out
}

func randomSpend(min, max float64) float64 {
	if max < min {
		min, max = max, min
	}
	return min + rand.Float64()*(max-min)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	return db, nil
}

func withLock(ctx context.Context, db *sql.DB, key int64, fn func(context.Context) error) error {
	if _, err := db.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, key); err != nil {
		return err
	}
	defer db.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, key)
	return fn(ctx)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
