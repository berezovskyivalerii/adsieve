package entity

import (
	"time"

	"github.com/google/uuid"
)

type Click struct {
	ID        int64     `json:"id" db:"id"`
	ClickID   string    `json:"click_id" db:"click_id"`
	AdID      int64     `json:"ad_id" db:"ad_id"`
	ClickedAt time.Time `json:"clicked_at" db:"clicked_at"`
	ClickRef  uuid.UUID `json:"click_ref" db:"click_ref"`
}

type ClickInput struct {
	ClickID   string    `json:"click_id"`
	AdID      int64     `json:"ad_id"`
	ClickedAt *int64    `json:"clicked_at"`
	ClickRef  uuid.UUID `json:"click_ref"`
}

func (in ClickInput) ParsedClickedAt() time.Time {
	if in.ClickedAt != nil && *in.ClickedAt > 0 {
		return time.Unix(*in.ClickedAt, 0).UTC()
	}
	return time.Now().UTC()
}
