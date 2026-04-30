package workflows

import (
	"testing"
	"time"
)

func TestRender_AllVars(t *testing.T) {
	tc := NewTriggerContext("/media/incoming/clip.mkv", time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC))
	cases := []struct {
		tmpl string
		want string
	}{
		{"{{path}}", "/media/incoming/clip.mkv"},
		{"{{dir}}", "/media/incoming"},
		{"{{filename}}", "clip.mkv"},
		{"{{stem}}", "clip"},
		{"{{ext}}", ".mkv"},
		{"{{date}}", "2026-04-30"},
		{"/out/{{stem}}.mp4", "/out/clip.mp4"},
		{"/out/{{ stem }}.mp4", "/out/clip.mp4"},   // whitespace tolerated
		{"/out/{{stem}}-{{date}}.mp4", "/out/clip-2026-04-30.mp4"},
	}
	for _, c := range cases {
		got := Render(c.tmpl, tc.Vars())
		if got != c.want {
			t.Errorf("Render(%q) = %q, want %q", c.tmpl, got, c.want)
		}
	}
}

func TestRender_UnknownKeyIsLeftLiteral(t *testing.T) {
	tc := NewTriggerContext("/x/y.mkv", time.Now())
	got := Render("/out/{{bogus}}-{{stem}}.mp4", tc.Vars())
	want := "/out/{{bogus}}-y.mp4"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestRender_NoInjection(t *testing.T) {
	// Filename that *looks* like a shell escape stays literal — Render
	// substitutes the value as-is; jobs.Create later passes it as a positional
	// argument to ffmpeg, never to a shell.
	tc := NewTriggerContext("/x/foo;rm -rf y.mkv", time.Now())
	got := Render("/out/{{filename}}", tc.Vars())
	want := "/out/foo;rm -rf y.mkv"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
