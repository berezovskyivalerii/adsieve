package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const lockKey int64 = 1003

func main() {
	dsn := getenv("DB_DSN", "postgres://user:pass@db:5432/adsieve?sslmode=disable")
	sqlFile := getenv("SQL_FILE", "/sql/aggregate_daily.sql")

	payload, err := os.ReadFile(sqlFile)
	if err != nil {
		log.Fatalf("read sql: %v", err)
	}

	db, err := openDB(dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := withLock(ctx, db, lockKey, func(ctx context.Context) error {
		tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return err
		}
		defer tx.Rollback()

		if _, err := tx.ExecContext(ctx, string(payload)); err != nil {
			return err
		}
		return tx.Commit()
	}); err != nil {
		log.Fatalf("aggregate failed: %v", err)
	}

	log.Printf("aggregate OK")
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
