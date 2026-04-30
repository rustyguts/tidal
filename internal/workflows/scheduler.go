package workflows

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/rustyguts/tidal/internal/domain"
)

// Scheduler runs one ticker goroutine per enabled workflow, plus a periodic
// reconciler that re-syncs from the DB so newly-enabled / newly-disabled
// workflows take effect without a server→worker notification channel.
type Scheduler struct {
	svc     *Service
	watcher *Watcher

	mu      sync.Mutex
	running map[domain.WorkflowID]context.CancelFunc

	rootCtx context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func NewScheduler(svc *Service, watcher *Watcher) *Scheduler {
	return &Scheduler{
		svc:     svc,
		watcher: watcher,
		running: make(map[domain.WorkflowID]context.CancelFunc),
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	s.rootCtx, s.cancel = context.WithCancel(ctx)
	return s.Sync(ctx)
}

// Sync reconciles the running goroutine set against enabled workflows in DB.
func (s *Scheduler) Sync(ctx context.Context) error {
	enabled, err := s.svc.ListEnabled(ctx)
	if err != nil {
		return err
	}
	want := make(map[domain.WorkflowID]domain.Workflow, len(enabled))
	for _, w := range enabled {
		want[w.ID] = w
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for id, cancel := range s.running {
		if _, ok := want[id]; !ok {
			cancel()
			delete(s.running, id)
		}
	}
	for id, w := range want {
		if _, ok := s.running[id]; ok {
			continue
		}
		gctx, cancel := context.WithCancel(s.rootCtx)
		s.running[id] = cancel
		s.wg.Add(1)
		go s.loop(gctx, w)
	}
	return nil
}

func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}

func (s *Scheduler) loop(ctx context.Context, w domain.Workflow) {
	defer s.wg.Done()
	interval := w.PollInterval()
	t := time.NewTimer(0)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := s.watcher.ScanOnce(ctx, w); err != nil {
				log.Warn().Str("workflow", w.Name).Err(err).Msg("scan tick")
			}
			t.Reset(interval)
		}
	}
}
