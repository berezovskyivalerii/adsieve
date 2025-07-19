package service

import (
	"context"

	"github.com/berezovskyivalerii/adsieve/internal/domain"
	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
)

type ConversionService struct {
	clickRepo domain.ClickRepository
	orderRepo domain.OrderRepository
}

func NewConversionService (c domain.ClickRepository, o domain.OrderRepository) *ConversionService{
	return &ConversionService{clickRepo: c, orderRepo: o}
}

func (s *ConversionService) Create(ctx context.Context, in entity.OrderInput) (int64, error){
    if _, err := s.clickRepo.ByClickID(ctx, in.ClickID); err == nil {
        return 0, errs.ErrClickNotFound
    }

    order := entity.Order{
        ClickID:    in.ClickID,
        OrderValue: in.OrderValue,
        OccurredAt: in.OccurredAtOrNow(),
    }
    id, err := s.orderRepo.Create(ctx, order)
    if err == errs.ErrOrderExists {
        return 0, err
    }
    return id, err     
}