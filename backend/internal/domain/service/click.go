package service

import (
	"context"

	"github.com/berezovskyivalerii/adsieve/internal/domain"
	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
)

type ClickService struct {
	repo domain.ClickRepository
}

func NewClickService(r domain.ClickRepository) *ClickService { return &ClickService{repo: r} }

func (s *ClickService) Click(ctx context.Context, clk entity.Click) (int64, error) {
	return s.repo.Click(ctx, clk)
}
