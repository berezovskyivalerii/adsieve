package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type AdsInsight struct {
	AdID        int64           `json:"ad_id"        db:"ad_id"`       
	InsightDate time.Time       `json:"insight_date" db:"insight_date"`
	Spend       decimal.Decimal `json:"spend"        db:"spend"`       
}
