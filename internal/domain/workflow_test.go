package domain

import (
	"testing"
	"time"
)

func TestWorkflow_PollInterval(t *testing.T) {
	tests := []struct {
		ms   int
		want time.Duration
	}{
		{ms: 0, want: 30 * time.Second},
		{ms: -1, want: 30 * time.Second},
		{ms: 1500, want: 1500 * time.Millisecond},
	}
	for _, tt := range tests {
		got := Workflow{PollIntervalMs: tt.ms}.PollInterval()
		if got != tt.want {
			t.Errorf("PollInterval(%d) = %v, want %v", tt.ms, got, tt.want)
		}
	}
}

func TestWorkflow_StableThreshold(t *testing.T) {
	tests := []struct {
		ms   int
		want time.Duration
	}{
		{ms: 0, want: 60 * time.Second},
		{ms: -100, want: 60 * time.Second},
		{ms: 1000, want: time.Second},
	}
	for _, tt := range tests {
		got := Workflow{StableThresholdMs: tt.ms}.StableThreshold()
		if got != tt.want {
			t.Errorf("StableThreshold(%d) = %v, want %v", tt.ms, got, tt.want)
		}
	}
}

func TestTrigger_Validate(t *testing.T) {
	tests := []struct {
		name    string
		t       Trigger
		wantErr bool
	}{
		{name: "valid", t: Trigger{Type: TriggerFileCreated, WatchDir: "/x", Glob: "*.mp4"}, wantErr: false},
		{name: "missing watchDir", t: Trigger{Type: TriggerFileCreated, Glob: "*.mp4"}, wantErr: true},
		{name: "missing glob", t: Trigger{Type: TriggerFileCreated, WatchDir: "/x"}, wantErr: true},
		{name: "unsupported type", t: Trigger{Type: "cron"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.t.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() = %v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestAction_Validate(t *testing.T) {
	tests := []struct {
		name    string
		a       Action
		wantErr bool
	}{
		{name: "valid", a: Action{Type: ActionEnqueueTranscode, PresetID: "abc"}, wantErr: false},
		{name: "missing presetId", a: Action{Type: ActionEnqueueTranscode}, wantErr: true},
		{name: "unsupported type", a: Action{Type: "webhook"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.a.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() = %v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestMarshalUnmarshalActions(t *testing.T) {
	in := []Action{{Type: ActionEnqueueTranscode, PresetID: "p1"}, {Type: ActionEnqueueTranscode, PresetID: "p2"}}
	raw, err := MarshalActions(in)
	if err != nil {
		t.Fatal(err)
	}
	out, err := UnmarshalActions(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 || out[0].PresetID != "p1" || out[1].PresetID != "p2" {
		t.Errorf("round-trip mismatch: %+v", out)
	}
}

func TestUnmarshalActions_empty(t *testing.T) {
	out, err := UnmarshalActions(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty, got %v", out)
	}
}

func TestUnmarshalActions_invalid(t *testing.T) {
	if _, err := UnmarshalActions([]byte("not json")); err == nil {
		t.Error("expected unmarshal error")
	}
}
