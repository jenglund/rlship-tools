package worker

import (
	"context"
	"log"
	"time"
)

// ShareCleanupService defines the interface needed for the worker
type ShareCleanupService interface {
	// CleanupExpiredShares removes expired shares
	CleanupExpiredShares() error
}

// ShareCleanupWorker is responsible for periodically cleaning up expired shares
type ShareCleanupWorker struct {
	service         ShareCleanupService
	cleanupInterval time.Duration
	ctx             context.Context
	cancelFunc      context.CancelFunc
}

// NewShareCleanupWorker creates a new worker for cleaning up expired shares
func NewShareCleanupWorker(service ShareCleanupService, cleanupInterval time.Duration) *ShareCleanupWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &ShareCleanupWorker{
		service:         service,
		cleanupInterval: cleanupInterval,
		ctx:             ctx,
		cancelFunc:      cancel,
	}
}

// Start begins the worker process
func (w *ShareCleanupWorker) Start() {
	log.Println("Starting share cleanup worker with interval:", w.cleanupInterval)

	// Run cleanup immediately on start
	if err := w.service.CleanupExpiredShares(); err != nil {
		log.Printf("Error during initial share cleanup: %v\n", err)
	} else {
		log.Println("Initial share cleanup completed successfully")
	}

	// Set up ticker for periodic cleanup
	ticker := time.NewTicker(w.cleanupInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("Running scheduled share cleanup...")
				if err := w.service.CleanupExpiredShares(); err != nil {
					log.Printf("Error during scheduled share cleanup: %v\n", err)
				} else {
					log.Println("Scheduled share cleanup completed successfully")
				}
			case <-w.ctx.Done():
				ticker.Stop()
				log.Println("Share cleanup worker stopped")
				return
			}
		}
	}()
}

// Stop halts the worker process
func (w *ShareCleanupWorker) Stop() {
	log.Println("Stopping share cleanup worker")
	w.cancelFunc()
}
