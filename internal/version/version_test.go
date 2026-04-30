package version

import (
	"strings"
	"testing"
)

func TestGet_Defaults(t *testing.T) {
	info := Get()
	if info.Version != "dev" {
		t.Errorf("Version = %q, want dev", info.Version)
	}
	if info.Commit != "unknown" {
		t.Errorf("Commit = %q, want unknown", info.Commit)
	}
	if info.Date != "unknown" {
		t.Errorf("Date = %q, want unknown", info.Date)
	}
}

func TestGet_Overridden(t *testing.T) {
	Version = "v1.0.0"
	Commit = "abc1234"
	Date = "2024-01-01T00:00:00Z"

	info := Get()
	if info.Version != "v1.0.0" {
		t.Errorf("Version = %q", info.Version)
	}
	if info.Commit != "abc1234" {
		t.Errorf("Commit = %q", info.Commit)
	}
	if !strings.HasPrefix(info.Date, "2024") {
		t.Errorf("Date = %q", info.Date)
	}
}
