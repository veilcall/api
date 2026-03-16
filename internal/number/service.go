package number

import (
	"context"
	"fmt"
	"time"
)

var planDuration = map[string]time.Duration{
	"24h": 24 * time.Hour,
	"7d":  7 * 24 * time.Hour,
	"30d": 30 * 24 * time.Hour,
}

var supportedCountries = map[string]bool{
	"US": true,
	"GB": true,
}

type NotifyHub interface {
	SendNotification(userID string, msg interface{})
}

type Service struct {
	repo   *Repository
	telnyx *TelnyxClient
	notify NotifyHub
}

func NewService(repo *Repository, telnyx *TelnyxClient, notify NotifyHub) *Service {
	return &Service{repo: repo, telnyx: telnyx, notify: notify}
}

// ProvisionForPayment is called by the payment poller when payment is confirmed.
func (s *Service) ProvisionForPayment(ctx context.Context, paymentID, userID, plan, country string) (string, error) {
	return s.provision(ctx, userID, plan, country)
}

func (s *Service) provision(ctx context.Context, userID, plan, country string) (string, error) {
	dur, ok := planDuration[plan]
	if !ok {
		return "", fmt.Errorf("invalid plan: %s", plan)
	}
	if !supportedCountries[country] {
		return "", fmt.Errorf("unsupported country: %s", country)
	}

	numbers, err := s.telnyx.SearchNumbers(ctx, country)
	if err != nil {
		return "", fmt.Errorf("search numbers: %w", err)
	}
	if len(numbers) == 0 {
		return "", fmt.Errorf("no numbers available for country: %s", country)
	}

	chosen := numbers[0]
	if err := s.telnyx.OrderNumber(ctx, chosen); err != nil {
		return "", fmt.Errorf("order number: %w", err)
	}

	n := &PhoneNumber{
		UserID:       userID,
		TelnyxNumber: chosen,
		Country:      country,
		Plan:         plan,
		ExpiresAt:    time.Now().Add(dur),
	}
	if err := s.repo.Create(ctx, n); err != nil {
		return "", err
	}
	return n.ID, nil
}

func (s *Service) ListNumbers(ctx context.Context, userID string) ([]*PhoneNumber, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) ReleaseNumber(ctx context.Context, id, userID string) error {
	n, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("number not found")
	}
	if n.Released {
		return fmt.Errorf("number already released")
	}
	if err := s.telnyx.ReleaseNumber(ctx, n.TelnyxNumber); err != nil {
		return fmt.Errorf("telnyx release: %w", err)
	}
	return s.repo.MarkReleased(ctx, id)
}

// GetByTelnyxNumber returns the userID of the number owner (for SMS routing).
func (s *Service) GetByTelnyxNumber(ctx context.Context, telnyxNumber string) (string, error) {
	n, err := s.repo.GetByTelnyxNumber(ctx, telnyxNumber)
	if err != nil {
		return "", err
	}
	return n.UserID, nil
}

// GetTelnyxNumber returns the E.164 number for a given numberID owned by userID.
func (s *Service) GetTelnyxNumber(ctx context.Context, numberID, userID string) (string, error) {
	n, err := s.repo.GetByID(ctx, numberID, userID)
	if err != nil {
		return "", err
	}
	return n.TelnyxNumber, nil
}
