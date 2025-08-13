package service

import (
	"context"

	"github.com/berezovskyivalerii/adsieve/internal/domain"
	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
)

type ConversionService struct {
	clickRepo      domain.ClickRepository
	conversionRepo domain.ConversionRepository
}

func NewConversionService(c domain.ClickRepository, conv domain.ConversionRepository) *ConversionService {
	return &ConversionService{clickRepo: c, conversionRepo: conv}
}

func (s *ConversionService) Create(ctx context.Context, in entity.ConversionInput) (int64, error) {
	click, err := s.clickRepo.ByClickID(ctx, in.ClickID) // Проверяем существует ли указаный клик
	if err != nil {
		if err == errs.ErrClickNotFound {
			return 0, errs.ErrClickNotFound
		}
		return 0, err
	}

	conv := entity.Conversion{
		AdID:        click.AdID, // cвязываем с таблицей ads
		ConvertedAt: in.ParsedConvertedAt(),
		Revenue:     in.Revenue,
		OrderID:     in.OrderID,
		ClickRef:    &click.ClickRef, // связь с помощью поля ClickRef с таблицей clicks
	}
	id, err := s.conversionRepo.Create(ctx, conv)
	if err == errs.ErrConversionExists {
		return 0, errs.ErrConversionExists
	}
	return id, err
}
