package queue

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/hibiken/asynq"

	"github.com/rustyguts/tidal/internal/domain"
)

// Task type names. Encoded into asynq.Task.Type.
const (
	TaskTypeTranscode = "tidal:transcode"
	TaskTypeScan      = "tidal:scan"
	TaskTypeCleanup   = "tidal:cleanup"
)

// Queue names. The asynq Server picks up tasks from these queues; the dispatcher
// pins each task type to its queue when enqueueing.
const (
	QueueTranscode = "transcode"
	QueueScan      = "scan"
	QueueDefault   = "default"
)

// TranscodePayload is the asynq payload for a single transcode job. The DB row
// is the source of truth; this carries just enough to look it up.
type TranscodePayload struct {
	JobID domain.JobID `json:"jobId"`
}

// ScanPayload triggers a single workflow scan tick.
type ScanPayload struct {
	WorkflowID domain.WorkflowID `json:"workflowId"`
}

// CleanupPayload prunes older job history.
type CleanupPayload struct {
	OlderThanHours int    `json:"olderThanHours"`
	Kind           string `json:"kind"`
}

func newTask(typ string, payload any, opts ...asynq.Option) (*asynq.Task, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal %s: %w", typ, err)
	}
	return asynq.NewTask(typ, body, opts...), nil
}

// ParseRedisURL converts a redis:// URL to asynq's RedisClientOpt.
func ParseRedisURL(rawURL string) (asynq.RedisClientOpt, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return asynq.RedisClientOpt{}, fmt.Errorf("parse redis url: %w", err)
	}
	opt := asynq.RedisClientOpt{Addr: u.Host}
	if u.User != nil {
		opt.Username = u.User.Username()
		if pw, ok := u.User.Password(); ok {
			opt.Password = pw
		}
	}
	if u.Path != "" && len(u.Path) > 1 {
		_, _ = fmt.Sscanf(u.Path[1:], "%d", &opt.DB)
	}
	return opt, nil
}
