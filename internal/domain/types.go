package domain

import (
	"time"

	"github.com/google/uuid"
)

type (
	PresetID     = uuid.UUID
	JobID        = uuid.UUID
	AutomationID = uuid.UUID
)

type JobStatus string

const (
	JobQueued     JobStatus = "queued"
	JobDispatched JobStatus = "dispatched"
	JobRunning    JobStatus = "running"
	JobCancelling JobStatus = "cancelling"
	JobSucceeded  JobStatus = "succeeded"
	JobFailed     JobStatus = "failed"
	JobCancelled  JobStatus = "cancelled"
)

func (s JobStatus) Terminal() bool {
	switch s {
	case JobSucceeded, JobFailed, JobCancelled:
		return true
	}
	return false
}

type Resolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Timestamps holds the standard created_at/updated_at pair.
type Timestamps struct {
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
