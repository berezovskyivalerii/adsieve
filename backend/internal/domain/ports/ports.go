// internal/domain/ports/ports.go
package ports

import (
	"context"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
)

type User interface {
	SignUp(ctx context.Context, inp entity.SignInput) (string, string, error)
	SignIn(ctx context.Context, in entity.SignInput) (string, string, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
}

type Click interface {
	Click(ctx context.Context, clk entity.ClickInput) (int64, error)
}

type Conversion interface {
	Create(ctx context.Context, in entity.ConversionInput) (int64, error)
}

type Metrics interface {
	Get(ctx context.Context, userID int64, f entity.MetricsFilter) ([]entity.DailyMetricDTO, error)
}

type Ads interface {
	List(ctx context.Context, userID int64, f entity.AdsFilter) (items []entity.AdDTO, total int, err error)
}

// === Google integrations ===
// Хранилище одноразовых состояний OAuth (state + PKCE)
type OAuthStateStore interface {
	Save(ctx context.Context, state, codeVerifier string, userID int64) error
	Consume(ctx context.Context, state string) (userID int64, codeVerifier string, err error) // atomically invalidate
}

// Хранилище refresh-токенов Google (с шифрованием внутри реализаций)
type TokenVault interface {
    SaveGoogleRefreshToken(ctx context.Context, userID int64, googleUserID, refreshToken, scope string) error
    LoadGoogleRefreshToken(ctx context.Context, userID int64) (googleUserID, refreshTokenEnc, scope string, err error)
    MarkNeedsConsent(ctx context.Context, userID int64, googleUserID string) error
}

// Клиент Google Ads для списка доступных аккаунтов и линковки
type GoogleAdsClient interface {
	ListAccessibleCustomers(ctx context.Context, userID int64) ([]string, error)
	LinkAccounts(ctx context.Context, userID int64, customerIDs []string) error
}
