package domain

import "time"

type Automation struct {
	ID           AutomationID `json:"id"`
	Name         string       `json:"name"`
	Enabled      bool         `json:"enabled"`
	WatchDir     string       `json:"watchDir"`
	Glob         string       `json:"glob"`
	PresetID     PresetID     `json:"presetId"`
	OutputDir    string       `json:"outputDir"`
	ArchiveDir   string       `json:"archiveDir"`
	PollInterval time.Duration `json:"-"`
	PollIntervalMs int        `json:"pollIntervalMs"`
	DebounceMs   int          `json:"debounceMs"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
}

type AutomationRun struct {
	ID           int64         `json:"id"`
	AutomationID AutomationID  `json:"automationId"`
	SourcePath   string        `json:"sourcePath"`
	JobID        *JobID        `json:"jobId,omitempty"`
	Outcome      string        `json:"outcome"`
	Message      string        `json:"message,omitempty"`
	OccurredAt   time.Time     `json:"occurredAt"`
}

const (
	OutcomeMatched     = "matched"
	OutcomeEnqueued    = "enqueued"
	OutcomeSkippedDupe = "skipped_dupe"
	OutcomeError       = "error"
	OutcomeArchived    = "archived"
)
