package domain

import "time"

type Job struct {
	ID             JobID          `json:"id"`
	AsynqID        string         `json:"asynqId,omitempty"`
	PresetID       PresetID       `json:"presetId"`
	SourcePath     string         `json:"sourcePath"`
	OutputPath     string         `json:"outputPath"`
	CachePath      string         `json:"cachePath,omitempty"`
	SourceMovePath string         `json:"sourceMovePath,omitempty"`
	Status         JobStatus      `json:"status"`
	K8sJobName     string         `json:"k8sJobName,omitempty"`
	WorkflowID     *WorkflowID    `json:"workflowId,omitempty"`
	Progress       FFmpegProgress `json:"progress"`
	Error          string         `json:"error,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
	StartedAt      *time.Time     `json:"startedAt,omitempty"`
	FinishedAt     *time.Time     `json:"finishedAt,omitempty"`
}

type FFmpegProgress struct {
	Frame     int64     `json:"frame"`
	FPS       float64   `json:"fps"`
	TimeMs    int64     `json:"timeMs"`
	Bitrate   int64     `json:"bitrate"`
	Speed     float64   `json:"speed"`
	SizeBytes int64     `json:"sizeBytes"`
	Percent   float64   `json:"percent"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type JobLog struct {
	JobID     JobID     `json:"jobId"`
	Seq       int64     `json:"seq"`
	Stream    string    `json:"stream"`
	Line      string    `json:"line"`
	EmittedAt time.Time `json:"emittedAt"`
}
