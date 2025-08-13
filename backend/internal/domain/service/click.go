package service

import (
	"context"

	"github.com/berezovskyivalerii/adsieve/internal/domain"
	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	"github.com/google/uuid"
)

type ClickService struct {
	repo domain.ClickRepository
}

func NewClickService(r domain.ClickRepository) *ClickService { return &ClickService{repo: r} }

func (s *ClickService) Click(ctx context.Context, in entity.ClickInput) (int64, error) {
	click := entity.Click{
		ClickID:   in.ClickID,
		AdID:      in.AdID,
		ClickedAt: in.ParsedClickedAt(),
		ClickRef:  uuid.New(),
	}
	return s.repo.Click(ctx, click)
}
