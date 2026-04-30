package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	"github.com/rustyguts/tidal/internal/domain"
)

// Client wraps asynq.Client + asynq.Inspector for the server-side enqueue
// and lifecycle operations (cancel, status).
type Client struct {
	cli       *asynq.Client
	inspector *asynq.Inspector
}

func NewClient(redisOpt asynq.RedisClientOpt) *Client {
	return &Client{
		cli:       asynq.NewClient(redisOpt),
		inspector: asynq.NewInspector(redisOpt),
	}
}

func (c *Client) Close() error {
	err := c.cli.Close()
	if ierr := c.inspector.Close(); ierr != nil && err == nil {
		err = ierr
	}
	return err
}

// EnqueueTranscode submits a transcode task to the transcode queue. Returns
// the asynq task ID so callers can persist it for cancellation.
func (c *Client) EnqueueTranscode(ctx context.Context, jobID domain.JobID) (string, error) {
	t, err := newTask(TaskTypeTranscode, TranscodePayload{JobID: jobID},
		asynq.Queue(QueueTranscode),
		asynq.MaxRetry(0),
		asynq.Timeout(24*time.Hour),
		asynq.Retention(7*24*time.Hour),
	)
	if err != nil {
		return "", err
	}
	info, err := c.cli.EnqueueContext(ctx, t)
	if err != nil {
		return "", fmt.Errorf("enqueue transcode: %w", err)
	}
	return info.ID, nil
}

// EnqueueScan submits a scan tick for an automation.
func (c *Client) EnqueueScan(ctx context.Context, automationID domain.AutomationID) (string, error) {
	t, err := newTask(TaskTypeScan, ScanPayload{AutomationID: automationID},
		asynq.Queue(QueueScan),
		asynq.MaxRetry(3),
		asynq.Timeout(5*time.Minute),
	)
	if err != nil {
		return "", err
	}
	info, err := c.cli.EnqueueContext(ctx, t)
	if err != nil {
		return "", fmt.Errorf("enqueue scan: %w", err)
	}
	return info.ID, nil
}

// CancelProcessing signals an in-flight task to abort. The asynq Inspector
// silently no-ops when the task is missing or already finished.
func (c *Client) CancelProcessing(asynqID string) error {
	if asynqID == "" {
		return nil
	}
	if err := c.inspector.CancelProcessing(asynqID); err != nil {
		return fmt.Errorf("cancel processing: %w", err)
	}
	return nil
}
