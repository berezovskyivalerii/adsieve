package entity

import "time"

type RefreshToken struct {
	TokenID      int64     `db:"token_id"`
	UserID       int64     `db:"user_id"`
	RefreshToken string    `db:"refresh_token"`
	ExpiresAt    time.Time `db:"expires_at"`
}
