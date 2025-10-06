package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type AdDailyMetric struct {
	AdID        int64           `json:"ad_id"        db:"ad_id"         validate:"required,gt=0"`
	Name        string          `json:"name"   		   db:"name"`
	Status      string          `json:"status" 		   db:"status"`
	MetricDate  time.Time       `json:"metric_date"  db:"metric_date"`
	Clicks      int             `json:"clicks"       db:"clicks"`
	Conversions int             `json:"conversions"  db:"conversions"`
	Revenue     decimal.Decimal `json:"revenue"      db:"revenue"`
	Spend       decimal.Decimal `json:"spend"        db:"spend"`
}

// CPA = spend / conversions
func (m AdDailyMetric) CPA() *decimal.Decimal {
	if m.Conversions == 0 {
		return nil
	}
	v := m.Spend.Div(decimal.NewFromInt(int64(m.Conversions)))
	return &v
}

// ROAS = revenue / spend
func (m AdDailyMetric) ROAS() *decimal.Decimal {
	if m.Spend.IsZero() {
		return nil
	}
	v := m.Revenue.Div(m.Spend)
	return &v
}

// DailyMetricDTO — формат JSON, который получает фронт.
// Денежные поля выводятся строкой, чтобы не терять точность в JS.
type DailyMetricDTO struct {
	AdID        int64  `json:"ad_id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Day         string `json:"day"` // YYYY-MM-DD
	Clicks      int    `json:"clicks"`
	Conversions int    `json:"conversions"`
	Revenue     string `json:"revenue"`
	Spend       string `json:"spend"`
	CPA         string `json:"cpa,omitempty"`
	ROAS        string `json:"roas,omitempty"`
}

// Фильтр, который принимает сервис
type MetricsFilter struct {
	AdIDs []int64   // пустой срез ⇒ все объявления пользователя
	From  time.Time // начало диапазона (включительно)
	To    time.Time // конец диапазона   (включительно)
}
