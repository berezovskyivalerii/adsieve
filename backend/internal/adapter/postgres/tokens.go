package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
)

type TokensRepo struct {
	db *sql.DB
}

func NewTokensRepo(db *sql.DB) *TokensRepo { return &TokensRepo{db: db} }

func (r *TokensRepo) Create(ctx context.Context, s entity.RefreshToken) error {
	const q = `INSERT INTO refresh_tokens (user_id, refresh_token, expires_at)
	           VALUES ($1, $2, $3)`

	_, err := r.db.ExecContext(ctx, q, s.UserID, s.RefreshToken, s.ExpiresAt)
	return err
}

// Get – читаем токен и сразу удаляем ВСЕ refresh-токены пользователя
func (r *TokensRepo) Get(ctx context.Context, rawToken string) (entity.RefreshToken, error) {
	var rt entity.RefreshToken

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return rt, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 1) Читаем запись
	const selectQ = `SELECT token_id, user_id, refresh_token, expires_at
	                 FROM   refresh_tokens
	                 WHERE  refresh_token = $1`
	err = tx.QueryRowContext(ctx, selectQ, rawToken).
		Scan(&rt.TokenID, &rt.UserID, &rt.RefreshToken, &rt.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return rt, sql.ErrNoRows
	}
	if err != nil {
		return rt, err
	}

	// 2) Удаляем все refresh-токены этого пользователя
	const deleteQ = `DELETE FROM refresh_tokens WHERE user_id = $1`
	if _, err = tx.ExecContext(ctx, deleteQ, rt.UserID); err != nil {
		return rt, err
	}

	err = tx.Commit()
	return rt, err
}
