package domain

import (
	"context"
	"time"

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
	Create(ctx context.Context, in entity.Conversion) (int64, error)
}

type MetricsRepository interface {
	List(ctx context.Context, adIDs []int64, from, to time.Time) ([]entity.AdDailyMetric, error)
}

type UserAdsRepo interface {
	IDsByUser(ctx context.Context, userID int64) ([]int64, error)
	Ensure(ctx context.Context, userID, adID int64) error
}

