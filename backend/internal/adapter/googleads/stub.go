package googleads

import (
	"context"
	"math/rand"
	"time"
)

// легкий мок: реализует и ports.GoogleAdsClient, и service.GoogleAdsCostStreamer
type StubRepo interface {
	LinkGoogleAccounts(ctx context.Context, userID int64, tokenOwnerGoogleUserID string, customerIDs []string) error
}

type Stub struct {
	repo StubRepo
}

func NewStub(repo StubRepo) *Stub { return &Stub{repo: repo} }

func (s *Stub) ListAccessibleCustomers(ctx context.Context, userID int64) ([]string, error) {
	// стаб отдает один "клиентский" аккаунт
	return []string{"999-000-1111"}, nil
}

func (s *Stub) LinkAccounts(ctx context.Context, userID int64, customerIDs []string) error {
	// владельца токена подставим фиктивно
	return s.repo.LinkGoogleAccounts(ctx, userID, "stub-google-user", customerIDs)
}

// используется сервисом синка
func (s *Stub) SyncCostsForDate(
	ctx context.Context, userID int64, customerID, yyyymmdd string,
	sink func(adID int64, date string, costMicros int64) error,
) error {
	// генерим 5 объявлений со "стоимостью" в микросах
	seed := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Unix() ^ int64(len(yyyymmdd)) ^ int64(len(customerID))
	r := rand.New(rand.NewSource(seed))
	for i := 0; i < 5; i++ {
		adID := 100000 + int64(i)
		costMicros := int64((r.Intn(9000) + 1000) * 1000) // от ~1 до ~10 у.е.
		if err := sink(adID, yyyymmdd, costMicros); err != nil {
			return err
		}
	}
	return nil
}
