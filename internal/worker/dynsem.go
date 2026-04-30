package worker

import (
	"context"
	"sync"
)

// DynSem is a semaphore whose capacity can change at runtime. New work waits
// when in-flight count >= max; shrinking max while above it is fine — already
// running tasks finish, but no new work is admitted until count drops.
type DynSem struct {
	mu   sync.Mutex
	cond *sync.Cond
	max  int
	cur  int
}

func NewDynSem(initial int) *DynSem {
	if initial <= 0 {
		initial = 1
	}
	s := &DynSem{max: initial}
	s.cond = sync.NewCond(&s.mu)
	return s
}

// Acquire blocks until a slot is free or ctx is done. Returns ctx.Err() on
// cancellation.
func (s *DynSem) Acquire(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// Wake the waiting goroutine if ctx fires.
	stop := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			s.cond.Broadcast()
		case <-stop:
		}
	}()
	defer close(stop)

	s.mu.Lock()
	defer s.mu.Unlock()
	for s.cur >= s.max {
		if err := ctx.Err(); err != nil {
			return err
		}
		s.cond.Wait()
	}
	s.cur++
	return nil
}

func (s *DynSem) Release() {
	s.mu.Lock()
	if s.cur > 0 {
		s.cur--
	}
	s.mu.Unlock()
	s.cond.Broadcast()
}

// SetMax updates the cap. Increasing wakes any waiters; decreasing only
// affects new acquires.
func (s *DynSem) SetMax(n int) {
	if n <= 0 {
		n = 1
	}
	s.mu.Lock()
	s.max = n
	s.mu.Unlock()
	s.cond.Broadcast()
}

func (s *DynSem) Max() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.max
}

func (s *DynSem) InFlight() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cur
}
