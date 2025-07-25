package entity

import (
	"github.com/shopspring/decimal"
	"time"

	"github.com/google/uuid"
)

type Conversion struct {
	ConversionID int64           `json:"conversion_id" db:"conversion_id"`
	AdID         int64           `json:"ad_id"         db:"ad_id"`
	ConvertedAt  time.Time       `json:"converted_at"  db:"converted_at"`
	Revenue      decimal.Decimal `json:"revenue"       db:"revenue"`
	OrderID      *string         `json:"order_id,omitempty"  db:"order_id"`
	ClickRef     *uuid.UUID      `json:"click_ref,omitempty" db:"click_ref"`
}

type ConversionInput struct {
	AdID        int64           `json:"ad_id"        validate:"required,min=1"`
	Revenue     decimal.Decimal `json:"revenue"      validate:"required"`
	OrderID     *string         `json:"order_id,omitempty"`
	ClickRef    *uuid.UUID      `json:"click_ref,omitempty"`
	ConvertedAt *int64          `json:"converted_at,omitempty"`
}

func (in ConversionInput) ParsedConvertedAt() time.Time {
	if in.ConvertedAt != nil && *in.ConvertedAt > 0 {
		return time.Unix(*in.ConvertedAt, 0).UTC()
	}
	return time.Now().UTC()
}
