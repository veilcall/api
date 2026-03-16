package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const telnyxMessagesURL = "https://api.telnyx.com/v2/messages"

// NumberLookup is implemented by number.Repository
type NumberLookup interface {
	GetByID(ctx context.Context, id, userID string) (interface{ GetTelnyxNumber() string }, error)
	GetByTelnyxNumber(ctx context.Context, telnyxNumber string) (interface{ GetUserID() string }, error)
}

// NumberService is implemented by number.Service
type NumberService interface {
	GetByTelnyxNumber(ctx context.Context, telnyxNumber string) (userID, numberID string, err error)
	GetNumberTelnyx(ctx context.Context, numberID, userID string) (string, error)
}

type Service struct {
	telnyxAPIKey string
	numSvc       NumberSvcAdapter
	hub          *Hub
	httpCli      *http.Client
}

// NumberSvcAdapter wraps number.Service for SMS use
type NumberSvcAdapter interface {
	GetByTelnyxNumber(ctx context.Context, telnyxNumber string) (userID string, err error)
	GetTelnyxNumber(ctx context.Context, numberID, userID string) (string, error)
}

func NewService(telnyxAPIKey string, numSvc NumberSvcAdapter, hub *Hub) *Service {
	return &Service{
		telnyxAPIKey: telnyxAPIKey,
		numSvc:       numSvc,
		hub:          hub,
		httpCli:      &http.Client{},
	}
}

func (s *Service) SendSMS(ctx context.Context, userID, fromNumberID, toE164, text string) error {
	telnyxNumber, err := s.numSvc.GetTelnyxNumber(ctx, fromNumberID, userID)
	if err != nil {
		return fmt.Errorf("number not found or unauthorized")
	}

	body, _ := json.Marshal(map[string]string{
		"from": telnyxNumber,
		"to":   toE164,
		"text": text,
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, telnyxMessagesURL, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+s.telnyxAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpCli.Do(req)
	if err != nil {
		return fmt.Errorf("telnyx send sms: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("telnyx send sms: status %d", resp.StatusCode)
	}
	return nil
}

// HandleInbound processes an inbound Telnyx webhook and pushes to WebSocket.
// Content is never persisted.
func (s *Service) HandleInbound(ctx context.Context, fromNumber, toNumber, text string) {
	userID, err := s.numSvc.GetByTelnyxNumber(ctx, toNumber)
	if err != nil {
		return
	}
	s.hub.SendNotification(userID, map[string]string{
		"type": "sms",
		"from": fromNumber,
		"text": text,
	})
	// text is not stored — intentional
}
