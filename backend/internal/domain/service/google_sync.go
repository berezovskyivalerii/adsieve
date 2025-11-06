package service

import (
	"context"
	"fmt"
)

type UserAdsRepo interface {
	Ensure(ctx context.Context, userID, adID int64) error
}

type GoogleAdsCostStreamer interface {
	SyncCostsForDate(ctx context.Context, userID int64, customerID, yyyymmdd string,
		sink func(adID int64, date string, costMicros int64) error) error
}

type GoogleAdAccountsRepo interface {
	GetAccountID(ctx context.Context, userID int64, platform, externalID string) (int64, error)
	UpsertAdIfMissing(ctx context.Context, accountID, adID int64) error
	UpsertSpend(ctx context.Context, adID int64, date string, costMicros int64) error
}

type GoogleSyncService struct {
	gads GoogleAdsCostStreamer
	repo GoogleAdAccountsRepo
	uads UserAdsRepo
}

func NewGoogleSync(gads GoogleAdsCostStreamer, repo GoogleAdAccountsRepo, uads UserAdsRepo) *GoogleSyncService {
	return &GoogleSyncService{gads: gads, repo: repo, uads: uads}
}

func (s *GoogleSyncService) SyncCostsForDate(ctx context.Context, userID int64, customerID, date string) error {
	accountID, err := s.repo.GetAccountID(ctx, userID, "google", customerID)
	if err != nil {
		return fmt.Errorf("lookup account_id: %w", err)
	}

	sink := func(adID int64, d string, costMicros int64) error {
		if err := s.repo.UpsertAdIfMissing(ctx, accountID, adID); err != nil {
			return fmt.Errorf("upsert ad %d: %w", adID, err)
		}
		// ключевая строка:
		if err := s.uads.Ensure(ctx, userID, adID); err != nil {
			return fmt.Errorf("link user->ad %d: %w", adID, err)
		}
		if err := s.repo.UpsertSpend(ctx, adID, d, costMicros); err != nil {
			return fmt.Errorf("upsert spend ad %d %s: %w", adID, d, err)
		}
		return nil
	}
	if err := s.gads.SyncCostsForDate(ctx, userID, customerID, date, sink); err != nil {
		return fmt.Errorf("google searchStream for %s %s: %w", customerID, date, err)
	}
	return nil
}
