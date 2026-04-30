package domain

import (
	"fmt"
	"strings"
)

type Preset struct {
	ID          PresetID     `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Builtin     bool         `json:"builtin"`
	Spec        PresetSpecV2 `json:"spec"`
	Timestamps
}

type PresetSpec struct {
	Container    string      `json:"container"`              // mp4|mkv|webm|mov
	VideoCodec   string      `json:"videoCodec"`             // libx264|libx265|libsvtav1|copy
	VideoPreset  string      `json:"videoPreset,omitempty"`  // ffmpeg -preset value
	CRF          int         `json:"crf"`
	PixelFormat  string      `json:"pixelFormat,omitempty"`
	AudioCodec   string      `json:"audioCodec,omitempty"`   // aac|libopus|copy
	AudioBitrate string      `json:"audioBitrate,omitempty"` // e.g. "192k"
	ExtraArgs    []string    `json:"extraArgs,omitempty"`
	OutputSuffix string      `json:"outputSuffix,omitempty"` // e.g. "_1080p"
	Resolution   *Resolution `json:"resolution,omitempty"`
}

var validContainers = map[string]struct{}{
	"mp4": {}, "mkv": {}, "webm": {}, "mov": {},
}

func (s PresetSpec) Validate() error {
	c := strings.ToLower(strings.TrimSpace(s.Container))
	if _, ok := validContainers[c]; !ok {
		return fmt.Errorf("container %q invalid (want mp4|mkv|webm|mov)", s.Container)
	}
	if strings.TrimSpace(s.VideoCodec) == "" {
		return fmt.Errorf("videoCodec required")
	}
	if s.CRF < 0 || s.CRF > 51 {
		return fmt.Errorf("crf %d out of range 0..51", s.CRF)
	}
	if s.Resolution != nil {
		if s.Resolution.Width <= 0 || s.Resolution.Width%2 != 0 {
			return fmt.Errorf("resolution.width must be positive and even")
		}
		if s.Resolution.Height <= 0 || s.Resolution.Height%2 != 0 {
			return fmt.Errorf("resolution.height must be positive and even")
		}
	}
	return nil
}

// ValidateName verifies the basic preset metadata (name only). Spec
// validation requires the catalog and lives on PresetSpecV2 / ValidateV2.
func (p Preset) ValidateName() error {
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("name required")
	}
	return nil
}
