package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	minConfirmations = 10
	xmrPriceBuffer   = 1.05 // 5% buffer
)

type Service struct {
	repo    *Repository
	monero  *MoneroClient
	prices  map[string]float64 // plan -> USD price
}

func NewService(repo *Repository, monero *MoneroClient, prices map[string]float64) *Service {
	return &Service{repo: repo, monero: monero, prices: prices}
}

type ReserveResult struct {
	PaymentID     string
	MoneroAddress string
	AmountXMR     float64
}

func (s *Service) Reserve(ctx context.Context, userID, plan, country string) (*ReserveResult, error) {
	usdPrice, ok := s.prices[plan]
	if !ok {
		return nil, fmt.Errorf("unknown plan: %s", plan)
	}

	xmrRate, err := fetchXMRRate(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch xmr rate: %w", err)
	}
	amountXMR := (usdPrice / xmrRate) * xmrPriceBuffer

	addrResult, err := s.monero.MakeIntegratedAddress(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("make integrated address: %w", err)
	}

	p := &Payment{
		UserID:        userID,
		MoneroAddress: addrResult.IntegratedAddress,
		MoneroAmount:  amountXMR,
		Plan:          plan,
		Country:       country,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	return &ReserveResult{
		PaymentID:     p.ID,
		MoneroAddress: addrResult.IntegratedAddress,
		AmountXMR:     amountXMR,
	}, nil
}

func (s *Service) GetStatus(ctx context.Context, paymentID, userID string) (*Payment, error) {
	return s.repo.GetByID(ctx, paymentID, userID)
}

type coinGeckoResponse struct {
	Monero struct {
		USD float64 `json:"usd"`
	} `json:"monero"`
}

func fetchXMRRate(ctx context.Context) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.coingecko.com/api/v3/simple/price?ids=monero&vs_currencies=usd", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var cg coinGeckoResponse
	if err := json.NewDecoder(resp.Body).Decode(&cg); err != nil {
		return 0, err
	}
	if cg.Monero.USD == 0 {
		return 0, fmt.Errorf("invalid xmr rate from coingecko")
	}
	return cg.Monero.USD, nil
}
