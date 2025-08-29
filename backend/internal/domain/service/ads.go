// internal/service/ads.go
package service

import (
	"context"
	"strings"

	"github.com/berezovskyivalerii/adsieve/internal/domain"
	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
)

// AdsService реализует бизнес-логику для GET /api/ads
type AdsService struct {
	repo domain.AdsRepository // ожидается метод: ListByUser(ctx, userID, f) ([]entity.Ad, int, error)
}

func NewAdsService(repo domain.AdsRepository) *AdsService {
	return &AdsService{repo: repo}
}

// List возвращает объявления пользователя с учётом фильтров/пагинации.
// Маппит сущности БД в DTO для API-ответа.
func (s *AdsService) List(ctx context.Context, userID int64, f entity.AdsFilter) ([]entity.AdDTO, int, error) {
	// лёгкая санитаризация sort (на случай, если хэндлер пропустил мусор)
	switch strings.ToLower(f.Sort) {
	case "name", "-name", "created_at", "-created_at":
	default:
		f.Sort = "name"
	}

	ads, total, err := s.repo.ListByUser(ctx, userID, f)
	if err != nil {
		return nil, 0, err
	}

	out := make([]entity.AdDTO, 0, len(ads))
	for _, a := range ads {
		out = append(out, entity.AdDTO{
			AdID:     a.AdID,
			Name:     a.Name,
			Status:   a.Status,
			Platform: a.Platform,
		})
	}
	return out, total, nil
}
