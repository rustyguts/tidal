package automations

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/rustyguts/tidal/internal/domain"
)

// Scheduler runs one ticker goroutine per enabled automation. Use Sync after
// any automation create/update/delete/enable/disable to reconcile.
type Scheduler struct {
	autos    *Service
	scanner  *Scanner

	mu      sync.Mutex
	running map[domain.AutomationID]context.CancelFunc

	rootCtx context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func NewScheduler(autos *Service, scanner *Scanner) *Scheduler {
	return &Scheduler{
		autos:   autos,
		scanner: scanner,
		running: make(map[domain.AutomationID]context.CancelFunc),
	}
}

// Start does an initial Sync and returns. Goroutines per automation tick on
// their own poll_interval. Stop drains them.
func (s *Scheduler) Start(ctx context.Context) error {
	s.rootCtx, s.cancel = context.WithCancel(ctx)
	return s.Sync(ctx)
}

// Sync starts/stops per-automation goroutines so the running set matches the
// enabled set in the DB.
func (s *Scheduler) Sync(ctx context.Context) error {
	enabled, err := s.autos.ListEnabled(ctx)
	if err != nil {
		return err
	}
	wantIDs := make(map[domain.AutomationID]domain.Automation, len(enabled))
	for _, a := range enabled {
		wantIDs[a.ID] = a
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Stop goroutines for automations that are no longer enabled.
	for id, cancel := range s.running {
		if _, ok := wantIDs[id]; !ok {
			cancel()
			delete(s.running, id)
		}
	}
	// Start goroutines for newly-enabled automations.
	for id, a := range wantIDs {
		if _, ok := s.running[id]; ok {
			continue
		}
		gctx, cancel := context.WithCancel(s.rootCtx)
		s.running[id] = cancel
		s.wg.Add(1)
		go s.loop(gctx, a)
	}
	return nil
}

func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}

func (s *Scheduler) loop(ctx context.Context, a domain.Automation) {
	defer s.wg.Done()
	interval := time.Duration(a.PollIntervalMs) * time.Millisecond
	if interval <= 0 {
		interval = 30 * time.Second
	}
	t := time.NewTimer(0)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := s.scanner.ScanOnce(ctx, a); err != nil {
				log.Warn().Str("automation", a.Name).Err(err).Msg("scan tick")
			}
			t.Reset(interval)
		}
	}
}
