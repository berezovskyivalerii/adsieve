package service

import (
	"context"
	"time"

	"github.com/berezovskyivalerii/adsieve/internal/domain"
	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
	"github.com/shopspring/decimal"
)

const maxRange = 90 * 24 * time.Hour

type MetricsService struct {
	metricsRepo domain.MetricsRepository
	userAdsRepo domain.UserAdsRepo
}

func NewMetricsService(m domain.MetricsRepository, a domain.UserAdsRepo) *MetricsService {
	return &MetricsService{metricsRepo: m, userAdsRepo: a}
}

func (s *MetricsService) Get(ctx context.Context, userID int64, f entity.MetricsFilter) ([]entity.DailyMetricDTO, error) {
	/* 1. Диапазон */
	if f.To.Before(f.From) || f.To.Sub(f.From) > maxRange {
		return nil, errs.ErrInvalidRange
	}

	/* 2. Доступные объявления пользователя */
	allowed, err := s.userAdsRepo.IDsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(allowed) == 0 {
		return nil, errs.ErrNoAdAccess
	}

	/* 3. Определяем итоговый scope ad_id */
	scope := intersect(allowed, f.AdIDs)
	if len(scope) == 0 {
		return nil, errs.ErrNoAdAccess
	}

	/* 4. Читаем агрегат */
	raw, err := s.metricsRepo.List(ctx, scope, f.From, f.To)
	if err != nil {
		return nil, err
	}

	/* 5. DTO + производные метрики */
	out := make([]entity.DailyMetricDTO, 0, len(raw))
	for _, m := range raw {
		dto := entity.DailyMetricDTO{
			AdID:        m.AdID,
			Name:        m.Name,
			Status:      m.Status,
			Day:         m.MetricDate.Format("2006-01-02"),
			Clicks:      m.Clicks,
			Conversions: m.Conversions,
			Revenue:     m.Revenue.StringFixed(2),
			Spend:       m.Spend.StringFixed(2),
		}

		if m.Conversions > 0 {
			dto.CPA = m.Spend.Div(decimal.NewFromInt(int64(m.Conversions))).StringFixed(2)
		}
		if !m.Spend.IsZero() {
			dto.ROAS = m.Revenue.Div(m.Spend).StringFixed(4)
		}
		out = append(out, dto)
	}

	return out, nil
}

func intersect(allowed, requested []int64) []int64 {
	if len(requested) == 0 {
		return allowed
	}
	set := make(map[int64]struct{}, len(allowed))
	for _, id := range allowed {
		set[id] = struct{}{}
	}
	out := make([]int64, 0, len(requested))
	for _, id := range requested {
		if _, ok := set[id]; ok {
			out = append(out, id)
		}
	}
	return out
}
