package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Encryptor interface {
	EncryptString(ctx context.Context, plain string) (string, error)
	// Если у тебя уже есть DecryptString — можно добавить сюда и использовать в TokenSource.
	// DecryptString(ctx context.Context, cipher string) (string, error)
}

type TokenVaultRepo struct {
	db  *sql.DB
	enc Encryptor
}

func NewTokenVault(db *sql.DB, enc Encryptor) *TokenVaultRepo {
	return &TokenVaultRepo{db: db, enc: enc}
}

// SaveGoogleRefreshToken — upsert шифрованного refresh-токена Google.
func (r *TokenVaultRepo) SaveGoogleRefreshToken(
	ctx context.Context,
	userID int64,
	googleUserID string,
	refreshToken string,
	scope string,
) error {
	encTok, err := r.enc.EncryptString(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("encrypt refresh: %w", err)
	}

	const q = `
INSERT INTO google_user_tokens (user_id, google_user_id, refresh_token_enc, refresh_token_scope, needs_consent, created_at, updated_at)
VALUES ($1, $2, $3, $4, FALSE, NOW(), NOW())
ON CONFLICT (user_id, google_user_id) DO UPDATE
SET refresh_token_enc   = EXCLUDED.refresh_token_enc,
    refresh_token_scope = EXCLUDED.refresh_token_scope,
    needs_consent       = FALSE,
    updated_at          = NOW()`

	if _, err = r.db.ExecContext(ctx, q, userID, googleUserID, encTok, scope); err != nil {
		return fmt.Errorf("save google refresh: %w", err)
	}
	return nil
}

// LoadGoogleRefreshToken — получить (google_user_id, refresh_token_enc, scope) по user_id.
// Возвращает sql.ErrNoRows, если токена нет.
func (r *TokenVaultRepo) LoadGoogleRefreshToken(
	ctx context.Context,
	userID int64,
) (googleUserID, refreshTokenEnc, scope string, err error) {
	const q = `
SELECT google_user_id, refresh_token_enc, refresh_token_scope
FROM google_user_tokens
WHERE user_id = $1 AND needs_consent = FALSE
LIMIT 1`
	err = r.db.QueryRowContext(ctx, q, userID).Scan(&googleUserID, &refreshTokenEnc, &scope)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", "", err
		}
		return "", "", "", fmt.Errorf("load google refresh: %w", err)
	}
	return
}

// MarkNeedsConsent — помечает запись как требующую повторного согласия.
// Используется при 401 от Google.
func (r *TokenVaultRepo) MarkNeedsConsent(
	ctx context.Context,
	userID int64,
	googleUserID string,
) error {
	const q = `
UPDATE google_user_tokens
SET needs_consent = TRUE, updated_at = NOW()
WHERE user_id = $1 AND google_user_id = $2`
	res, err := r.db.ExecContext(ctx, q, userID, googleUserID)
	if err != nil {
		return fmt.Errorf("mark needs consent: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		// Не критично, но полезно знать, что записи не было
		return sql.ErrNoRows
	}
	return nil
}
