package entity

import "time"

type AdAccount struct {
	AccountID         int64     `json:"account_id"   db:"account_id"`
	UserID            int64     `json:"user_id"      db:"user_id"`
	Platform          string    `json:"platform"     db:"platform"`
	ExternalAccountID string    `json:"external_account_id" db:"external_account_id"` // ID на рекламной платформе
	AccessToken       string    `json:"access_token" db:"access_token"` // long-lived token
	CreatedAt         time.Time `json:"created_at"   db:"created_at"`
}
