package entity

import "time"

type Click struct {
	ID        int64     `json:"id" db:"id"`
	ClickID   string    `json:"click_id" db:"click_id"`
	AdID      int64     `json:"ad_id" db:"ad_id"`
	OccurredAt time.Time `json:"occured_at" db:"occured_at"`
}
