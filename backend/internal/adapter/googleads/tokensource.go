package googleads

import (
	"context"
	"time"

	"golang.org/x/oauth2"
)

type Vault interface {
	LoadRefreshToken(ctx context.Context, userID int64) (googleUserID, refreshTokenEnc, scope string, err error)
	Decrypt(s string) (string, error)
	MarkNeedsConsent(ctx context.Context, userID int64, googleUserID string) error
}

type OAuthCfg interface {
	ExchangeRefresh(ctx context.Context, refresh string) (*oauth2.Token, error)
}

type TS struct {
	v  Vault
	oc OAuthCfg
}

func NewTokenSource(v Vault, oc OAuthCfg) *TS { return &TS{v: v, oc: oc} }

func (t *TS) Token(ctx context.Context, userID int64) (string, string, error) {
	googleUID, refreshEnc, _, err := t.v.LoadRefreshToken(ctx, userID)
	if err != nil {
		return "", "", err
	}
	refresh, err := t.v.Decrypt(refreshEnc)
	if err != nil {
		return "", "", err
	}
	tok, err := t.oc.ExchangeRefresh(ctx, refresh)
	if err != nil {
		return "", "", err
	}
	if !tok.Expiry.IsZero() && time.Until(tok.Expiry) < time.Minute {
		// форс обновление, если вдруг на грани (опционально)
	}
	return tok.AccessToken, googleUID, nil
}

func (t *TS) MarkNeedsConsent(ctx context.Context, userID int64, googleUserID string) error {
	return t.v.MarkNeedsConsent(ctx, userID, googleUserID)
}
