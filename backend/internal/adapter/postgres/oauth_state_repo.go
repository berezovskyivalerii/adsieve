package postgres

import (
	"context"
	"database/sql"
	"time"
)

type OAuthStateRepo struct {
	db  *sql.DB
	ttl time.Duration
}

func NewOAuthStateRepo(db *sql.DB, ttl time.Duration) *OAuthStateRepo {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &OAuthStateRepo{db: db, ttl: ttl}
}

// Save(state, verifier, userID) с TTL
func (r *OAuthStateRepo) Save(ctx context.Context, state, codeVerifier string, userID int64) error {
	const q = `
  	INSERT INTO oauth_states (state, code_verifier, user_id, expires_at)
  	VALUES ($1, $2, $3, $4)
 	  ON CONFLICT (state) DO UPDATE
  	SET code_verifier = EXCLUDED.code_verifier,
    	  user_id       = EXCLUDED.user_id,
  	    expires_at    = EXCLUDED.expires_at;`

	expiresAt := time.Now().UTC().Add(r.ttl)
	_, err := r.db.ExecContext(ctx, q, state, codeVerifier, userID, expiresAt)
	return err
}

// Consume атомарно достаёт и удаляет запись (провалится при истёкшем TTL)
func (r *OAuthStateRepo) Consume(ctx context.Context, state string) (int64, string, error) {
	const q = `
		DELETE FROM oauth_states
		WHERE state = $1 AND expires_at > NOW()
		RETURNING user_id, code_verifier`
	var userID int64
	var verifier string
	err := r.db.QueryRowContext(ctx, q, state).Scan(&userID, &verifier)
	if err != nil {
		return 0, "", err
	}
	return userID, verifier, nil
}
