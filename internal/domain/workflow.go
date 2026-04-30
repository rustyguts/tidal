package domain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type WorkflowID = uuid.UUID

type Workflow struct {
	ID                WorkflowID `json:"id"`
	Name              string     `json:"name"`
	Enabled           bool       `json:"enabled"`
	Trigger           Trigger    `json:"trigger"`
	Actions           []Action   `json:"actions"`
	PollIntervalMs    int        `json:"pollIntervalMs"`
	StableThresholdMs int        `json:"stableThresholdMs"`
	RunsCount         int64      `json:"runsCount"`
	SuccessCount      int64      `json:"successCount"`
	LastRunAt         *time.Time `json:"lastRunAt,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

// PollInterval is a derived helper.
func (w Workflow) PollInterval() time.Duration {
	if w.PollIntervalMs <= 0 {
		return 30 * time.Second
	}
	return time.Duration(w.PollIntervalMs) * time.Millisecond
}

func (w Workflow) StableThreshold() time.Duration {
	if w.StableThresholdMs <= 0 {
		return 60 * time.Second
	}
	return time.Duration(w.StableThresholdMs) * time.Millisecond
}

// Trigger is a tagged union — `Type` discriminates fields. Currently only
// `file_created` is supported. JSON round-trips via custom
// Marshaler/Unmarshaler.
type Trigger struct {
	Type     string `json:"type"`
	WatchDir string `json:"watchDir,omitempty"`
	Glob     string `json:"glob,omitempty"`
}

const (
	TriggerFileCreated = "file_created"
)

func (t Trigger) Validate() error {
	switch t.Type {
	case TriggerFileCreated:
		if t.WatchDir == "" {
			return fmt.Errorf("trigger.watchDir required")
		}
		if t.Glob == "" {
			return fmt.Errorf("trigger.glob required")
		}
		return nil
	default:
		return fmt.Errorf("unsupported trigger type %q", t.Type)
	}
}

// Action is a tagged union — `Type` discriminates the per-type fields.
// Currently only `enqueue_transcode` is supported.
type Action struct {
	Type           string `json:"type"`
	PresetID       string `json:"presetId,omitempty"`
	OutputPath     string `json:"outputPath,omitempty"`
	SourceMovePath string `json:"sourceMovePath,omitempty"`
	CachePath      string `json:"cachePath,omitempty"`
}

const (
	ActionEnqueueTranscode = "enqueue_transcode"
)

func (a Action) Validate() error {
	switch a.Type {
	case ActionEnqueueTranscode:
		if a.PresetID == "" {
			return fmt.Errorf("action.presetId required")
		}
		return nil
	default:
		return fmt.Errorf("unsupported action type %q", a.Type)
	}
}

// MarshalActions / UnmarshalActions handle the JSONB column round-trip.
func MarshalActions(actions []Action) ([]byte, error) { return json.Marshal(actions) }

func UnmarshalActions(data []byte) ([]Action, error) {
	var out []Action
	if len(data) == 0 {
		return out, nil
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

type WorkflowRun struct {
	ID          int64      `json:"id"`
	WorkflowID  WorkflowID `json:"workflowId"`
	SourcePath  string     `json:"sourcePath"`
	JobID       *JobID     `json:"jobId,omitempty"`
	Outcome     string     `json:"outcome"`
	Message     string     `json:"message,omitempty"`
	OccurredAt  time.Time  `json:"occurredAt"`
}

const (
	WorkflowOutcomeTriggered   = "triggered"
	WorkflowOutcomeEnqueued    = "enqueued"
	WorkflowOutcomeSkippedDupe = "skipped_dupe"
	WorkflowOutcomeError       = "error"
	WorkflowOutcomeSucceeded   = "succeeded"
	WorkflowOutcomeFailed      = "failed"
)
