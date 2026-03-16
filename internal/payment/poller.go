package payment

import (
	"context"
	"log"
	"time"
)

// NumberProvisioner is implemented by number.Service
type NumberProvisioner interface {
	ProvisionForPayment(ctx context.Context, paymentID, userID, plan, country string) (string, error)
}

type Poller struct {
	repo        *Repository
	monero      *MoneroClient
	provisioner NumberProvisioner
}

func NewPoller(repo *Repository, monero *MoneroClient, provisioner NumberProvisioner) *Poller {
	return &Poller{repo: repo, monero: monero, provisioner: provisioner}
}

func (p *Poller) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.poll(ctx)
		}
	}
}

func (p *Poller) poll(ctx context.Context) {
	// expire old pending payments
	if err := p.repo.ExpireOld(ctx); err != nil {
		log.Printf("poller expire old: %v", err)
	}

	pending, err := p.repo.ListPending(ctx)
	if err != nil {
		log.Printf("poller list pending: %v", err)
		return
	}
	if len(pending) == 0 {
		return
	}

	transfers, err := p.monero.GetTransfers(ctx)
	if err != nil {
		log.Printf("poller get transfers: %v", err)
		return
	}

	// Build a lookup: address -> transfer (only confirmed ones)
	confirmed := map[string]Transfer{}
	for _, t := range transfers {
		if t.Confirmations >= minConfirmations {
			confirmed[t.Address] = t
		}
	}

	for _, payment := range pending {
		t, ok := confirmed[payment.MoneroAddress]
		if !ok {
			continue
		}

		// Verify amount (picomonero: 1 XMR = 1e12)
		expected := uint64(payment.MoneroAmount * 1e12)
		if t.Amount < expected {
			log.Printf("poller payment %s: insufficient amount %d < %d", payment.ID, t.Amount, expected)
			continue
		}

		numberID, err := p.provisioner.ProvisionForPayment(ctx, payment.ID, payment.UserID, payment.Plan, payment.Country)
		if err != nil {
			log.Printf("poller provision for payment %s: %v", payment.ID, err)
			continue
		}
		if err := p.repo.Confirm(ctx, payment.ID, numberID); err != nil {
			log.Printf("poller confirm payment %s: %v", payment.ID, err)
		}
	}
}
