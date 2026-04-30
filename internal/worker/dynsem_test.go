package worker

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDynSemRespectsCap(t *testing.T) {
	t.Parallel()
	s := NewDynSem(2)
	var inFlight, peak int32
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.Acquire(context.Background()); err != nil {
				t.Errorf("acquire: %v", err)
				return
			}
			cur := atomic.AddInt32(&inFlight, 1)
			for {
				p := atomic.LoadInt32(&peak)
				if cur <= p || atomic.CompareAndSwapInt32(&peak, p, cur) {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt32(&inFlight, -1)
			s.Release()
		}()
	}
	wg.Wait()
	if peak > 2 {
		t.Fatalf("peak in-flight %d > 2", peak)
	}
}

func TestDynSemSetMaxWakesWaiters(t *testing.T) {
	t.Parallel()
	s := NewDynSem(1)
	if err := s.Acquire(context.Background()); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		_ = s.Acquire(context.Background())
		close(done)
	}()
	// Without SetMax this would block forever (cap=1, in-flight=1).
	time.Sleep(20 * time.Millisecond)
	select {
	case <-done:
		t.Fatal("acquired before SetMax")
	default:
	}
	s.SetMax(2)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("SetMax did not wake waiter")
	}
}

func TestDynSemAcquireCtxCancel(t *testing.T) {
	t.Parallel()
	s := NewDynSem(1)
	if err := s.Acquire(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := s.Acquire(ctx); err == nil {
		t.Fatal("expected ctx error")
	}
}
