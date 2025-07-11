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
    if _, err := s.clickRepo.ByID(ctx, in.ClickID); err == nil {
        return 0, errs.ErrClickNotFound
    }

    // 2. Пытаемся вставить заказ
    order := entity.Order{
        ClickID:    in.ClickID,
        OrderValue: in.OrderValue,
        OccurredAt: in.OccurredAtOrNow(),
    }
    id, err := s.orderRepo.Create(ctx, order)
    if err == errs.ErrOrderExists {              // дубликат заказа
        return 0, err                           // 409
    }
    return id, err     
}