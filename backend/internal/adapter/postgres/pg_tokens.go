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

// -----------------------------------------------------------------------------
// Create – сохраняем refresh-токен (plain database/sql)
// -----------------------------------------------------------------------------
func (r *TokensRepo) Create(ctx context.Context, s entity.RefreshSession) error {
	const q = `INSERT INTO refresh_tokens (user_id, token, expires_at)
	           VALUES ($1, $2, $3)`

	_, err := r.db.ExecContext(ctx, q, s.UserID, s.Token, s.ExpiresAt)
	return err
}

// -----------------------------------------------------------------------------
// Get – читаем токен и сразу удаляем все refresh-токены пользователя
// (single-use + one-session policy)
// -----------------------------------------------------------------------------
func (r *TokensRepo) Get(ctx context.Context, rawToken string) (entity.RefreshSession, error) {
	var rs entity.RefreshSession

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return rs, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 1) Читаем запись
	const selectQ = `SELECT id, user_id, token, expires_at
	                 FROM   refresh_tokens
	                 WHERE  token = $1`
	err = tx.QueryRowContext(ctx, selectQ, rawToken).
		Scan(&rs.ID, &rs.UserID, &rs.Token, &rs.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return rs, sql.ErrNoRows
	}
	if err != nil {
		return rs, err
	}

	// 2) Удаляем все refresh-токены этого пользователя
	const deleteQ = `DELETE FROM refresh_tokens WHERE user_id = $1`
	if _, err = tx.ExecContext(ctx, deleteQ, rs.UserID); err != nil {
		return rs, err
	}

	err = tx.Commit()
	return rs, err
}
