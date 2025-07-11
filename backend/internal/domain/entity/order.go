package entity

import (
	"github.com/shopspring/decimal"
	"time"
)

type Order struct {
	ID         int64           `db:"id"`
	ClickID    string          `db:"click_id"`
	OrderValue decimal.Decimal `db:"order_value"`
	OccurredAt time.Time       `db:"occurred_at"`
}

type OrderInput struct {
	ClickID    string          `json:"click_id"`
	OrderValue decimal.Decimal `json:"order_value"`
	OccurredAt time.Time       `json:"occurred_at"`
}

func (e OrderInput) OccurredAtOrNow() time.Time {
	if e.OccurredAt.IsZero() {
		return time.Now().UTC()
	}
	return e.OccurredAt
}