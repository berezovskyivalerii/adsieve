package domain

import (
	"context"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
)

type AuthRepository interface {
	ByEmail(ctx context.Context, email string) (entity.User, error)
	CreateUser(ctx context.Context, user entity.User) (int64, error)
}