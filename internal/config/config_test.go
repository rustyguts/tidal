package config

import (
	"testing"
)

func TestValidate_DispatcherLocal(t *testing.T) {
	c := Config{Dispatcher: "local"}
	if err := c.validate(); err != nil {
		t.Errorf("local dispatcher should be valid: %v", err)
	}
}

func TestValidate_DispatcherK8s(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "k8s with required fields",
			cfg:     Config{Dispatcher: "k8s", JobImage: "tidal:latest"},
			wantErr: false,
		},
		{
			name:    "k8s missing job image",
			cfg:     Config{Dispatcher: "k8s", JobImage: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidate_InvalidDispatcher(t *testing.T) {
	c := Config{Dispatcher: "invalid"}
	if err := c.validate(); err == nil {
		t.Error("invalid dispatcher should error")
	}
}

func TestStringSlice_Decode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty string", "", nil},
		{"single value", "/media", []string{"/media"}},
		{"multiple values", "/media,/data", []string{"/media", "/data"}},
		{"with spaces", "/media, /data", []string{"/media", "/data"}},
		{"trailing comma", "/media,", []string{"/media"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s StringSlice
			if err := s.Decode(tt.input); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if len(s) != len(tt.want) {
				t.Fatalf("got len=%d, want len=%d", len(s), len(tt.want))
			}
			for i := range tt.want {
				if s[i] != tt.want[i] {
					t.Errorf("s[%d] = %q, want %q", i, s[i], tt.want[i])
				}
			}
		})
	}
}
