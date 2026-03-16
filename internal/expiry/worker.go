package expiry

import (
	"context"
	"log"
	"time"

	"github.com/dfhgiudhv/privatecall/internal/number"
)

type NotifyHub interface {
	SendNotification(userID string, msg interface{})
}

type Worker struct {
	numRepo *number.Repository
	telnyx  *number.TelnyxClient
	notify  NotifyHub
}

func NewWorker(numRepo *number.Repository, telnyx *number.TelnyxClient, notify NotifyHub) *Worker {
	return &Worker{numRepo: numRepo, telnyx: telnyx, notify: notify}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.run(ctx)
		}
	}
}

func (w *Worker) run(ctx context.Context) {
	expired, err := w.numRepo.ListExpired(ctx)
	if err != nil {
		log.Printf("expiry worker list expired: %v", err)
		return
	}
	for _, n := range expired {
		if err := w.telnyx.ReleaseNumber(ctx, n.TelnyxNumber); err != nil {
			log.Printf("expiry worker release %s: %v", n.TelnyxNumber, err)
			// Still mark as released to avoid repeated attempts
		}
		if err := w.numRepo.MarkReleased(ctx, n.ID); err != nil {
			log.Printf("expiry worker mark released %s: %v", n.ID, err)
			continue
		}
		w.notify.SendNotification(n.UserID, map[string]string{
			"type":   "number_expired",
			"number": n.TelnyxNumber,
		})
	}
}
