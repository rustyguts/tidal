package workflows

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/rustyguts/tidal/internal/domain"
)

type observation struct {
	size       int64
	mtime      time.Time
	observedAt time.Time
}

// Watcher remembers per-file observations across scan ticks so the executor
// only fires once a file has been size+mtime stable for stable_threshold.
type Watcher struct {
	exec *Executor

	mu  sync.Mutex
	obs map[string]observation
}

func NewWatcher(exec *Executor) *Watcher {
	return &Watcher{exec: exec, obs: make(map[string]observation)}
}

// ScanOnce enumerates the workflow's watch dir, applies stable-file rules,
// fires the executor for each newly-stable match, and prunes very old
// observation entries to bound memory.
func (w *Watcher) ScanOnce(ctx context.Context, wf domain.Workflow) error {
	if wf.Trigger.Type != domain.TriggerFileCreated {
		return nil
	}
	dir := wf.Trigger.WatchDir
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	stable := wf.StableThreshold()
	now := time.Now()

	w.mu.Lock()
	defer w.mu.Unlock()

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		match, err := filepath.Match(wf.Trigger.Glob, e.Name())
		if err != nil || !match {
			continue
		}
		full := filepath.Join(dir, e.Name())
		fi, err := e.Info()
		if err != nil {
			continue
		}

		prev, seen := w.obs[full]
		if seen && prev.size == fi.Size() && prev.mtime.Equal(fi.ModTime()) &&
			now.Sub(prev.observedAt) >= stable && now.Sub(fi.ModTime()) >= stable {
			// Stable — fire.
			delete(w.obs, full)
			tc := NewTriggerContext(full, now)
			if err := w.exec.Execute(ctx, wf, tc); err != nil {
				log.Warn().Err(err).Str("workflow", wf.Name).Str("source", full).Msg("execute")
			}
			continue
		}
		// First observation OR file still changing OR not yet stable long enough.
		w.obs[full] = observation{
			size:       fi.Size(),
			mtime:      fi.ModTime(),
			observedAt: pickObservedAt(prev, seen, fi),
		}
	}

	w.gc(now)
	return nil
}

// pickObservedAt resets the stability timer when the file changes; otherwise
// keeps the prior observed_at so the threshold counts from the first stable
// observation, not from the latest.
func pickObservedAt(prev observation, seen bool, fi os.FileInfo) time.Time {
	if seen && prev.size == fi.Size() && prev.mtime.Equal(fi.ModTime()) {
		return prev.observedAt
	}
	return time.Now()
}

// gc evicts entries older than 24h to keep the observation map bounded for
// long-lived watchers over churning directories.
func (w *Watcher) gc(now time.Time) {
	const ttl = 24 * time.Hour
	for path, o := range w.obs {
		if now.Sub(o.observedAt) > ttl {
			delete(w.obs, path)
		}
	}
}
