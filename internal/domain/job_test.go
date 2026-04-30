package domain

import "testing"

func TestJobStatus_Terminal(t *testing.T) {
	tests := []struct {
		status JobStatus
		want   bool
	}{
		{JobQueued, false},
		{JobDispatched, false},
		{JobRunning, false},
		{JobCancelling, false},
		{JobSucceeded, true},
		{JobFailed, true},
		{JobCancelled, true},
		{JobStatus("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.Terminal(); got != tt.want {
				t.Errorf("Terminal() = %v, want %v", got, tt.want)
			}
		})
	}
}