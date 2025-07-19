package domain

import (
	"context"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
)

type ClickRepository interface {
	Click(ctx context.Context, clk entity.Click) (int64, error)
	ByClickID(ctx context.Context, id string) (entity.Click, error)
}

type AuthRepository interface {
	ByEmail(ctx context.Context, email string) (entity.User, error)
	CreateUser(ctx context.Context, user entity.User) (int64, error)
}

type OrderRepository interface {
	Create(ctx context.Context, in entity.Order) (int64, error)
}

