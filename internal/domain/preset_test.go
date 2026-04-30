package domain

import (
	"testing"
)

func TestPresetSpec_Validate(t *testing.T) {
	tests := []struct {
		name    string
		spec    PresetSpec
		wantErr bool
	}{
		{
			name: "valid minimal",
			spec: PresetSpec{
				Container:  "mp4",
				VideoCodec: "libx264",
				CRF:        23,
			},
			wantErr: false,
		},
		{
			name: "invalid container",
			spec: PresetSpec{
				Container:  "avi",
				VideoCodec: "libx264",
				CRF:        23,
			},
			wantErr: true,
		},
		{
			name: "empty container",
			spec: PresetSpec{
				Container:  "",
				VideoCodec: "libx264",
				CRF:        23,
			},
			wantErr: true,
		},
		{
			name: "empty video codec",
			spec: PresetSpec{
				Container:  "mkv",
				VideoCodec: "",
				CRF:        23,
			},
			wantErr: true,
		},
		{
			name: "crf below range",
			spec: PresetSpec{
				Container:  "mp4",
				VideoCodec: "libx265",
				CRF:        -1,
			},
			wantErr: true,
		},
		{
			name: "crf above range",
			spec: PresetSpec{
				Container:  "mp4",
				VideoCodec: "libx264",
				CRF:        52,
			},
			wantErr: true,
		},
		{
			name: "crf valid boundary 0",
			spec: PresetSpec{
				Container:  "webm",
				VideoCodec: "libsvtav1",
				CRF:        0,
			},
			wantErr: false,
		},
		{
			name: "crf valid boundary 51",
			spec: PresetSpec{
				Container:  "mov",
				VideoCodec: "libx264",
				CRF:        51,
			},
			wantErr: false,
		},
		{
			name: "case insensitive container",
			spec: PresetSpec{
				Container:  "MP4",
				VideoCodec: "libx264",
				CRF:        23,
			},
			wantErr: false,
		},
		{
			name: "resolution valid",
			spec: PresetSpec{
				Container:  "mp4",
				VideoCodec: "libx264",
				CRF:        23,
				Resolution: &Resolution{Width: 1920, Height: 1080},
			},
			wantErr: false,
		},
		{
			name: "resolution odd width",
			spec: PresetSpec{
				Container:  "mp4",
				VideoCodec: "libx264",
				CRF:        23,
				Resolution: &Resolution{Width: 1921, Height: 1080},
			},
			wantErr: true,
		},
		{
			name: "resolution odd height",
			spec: PresetSpec{
				Container:  "mp4",
				VideoCodec: "libx264",
				CRF:        23,
				Resolution: &Resolution{Width: 1920, Height: 1081},
			},
			wantErr: true,
		},
		{
			name: "resolution zero width",
			spec: PresetSpec{
				Container:  "mp4",
				VideoCodec: "libx264",
				CRF:        23,
				Resolution: &Resolution{Width: 0, Height: 1080},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestPreset_Validate(t *testing.T) {
	tests := []struct {
		name    string
		preset  Preset
		wantErr bool
	}{
		{
			name: "valid preset",
			preset: Preset{
				Name:        "test",
				Description: "a test preset",
				Spec: PresetSpec{
					Container:  "mp4",
					VideoCodec: "libx264",
					CRF:        23,
				},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			preset: Preset{
				Name: "",
				Spec: PresetSpec{
					Container:  "mp4",
					VideoCodec: "libx264",
					CRF:        23,
				},
			},
			wantErr: true,
		},
		{
			name: "spec failure propagates",
			preset: Preset{
				Name: "bad",
				Spec: PresetSpec{
					Container:  "avi",
					VideoCodec: "libx264",
					CRF:        23,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.preset.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
