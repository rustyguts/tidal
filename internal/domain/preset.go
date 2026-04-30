package domain

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Preset struct {
	ID             PresetID   `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Builtin        bool       `json:"builtin"`
	Spec           PresetSpec `json:"spec"`
	OutputPath     string     `json:"outputPath,omitempty"`
	CachePath      string     `json:"cachePath,omitempty"`
	SourceMovePath string     `json:"sourceMovePath,omitempty"`
	Timestamps
}

// ValidateName verifies the basic preset metadata (name only). Spec
// validation is catalog-driven and lives on PresetSpec / Validate.
func (p Preset) ValidateName() error {
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("name required")
	}
	return nil
}

// PresetSpec is the structured ffmpeg-encode spec. Sections are independently
// composed by the builder; cross-field validation lives in preset_validate.go.
type PresetSpec struct {
	Container    ContainerSpec `json:"container"`
	Input        InputSpec     `json:"input,omitempty"`
	Hwaccel      *HwaccelSpec  `json:"hwaccel,omitempty"`
	Video        VideoSpec     `json:"video"`
	Audio        AudioSpec     `json:"audio"`
	Filters      FilterChain   `json:"filters,omitempty"`
	Mapping      MappingSpec   `json:"mapping,omitempty"`
	Subtitles    SubtitleSpec  `json:"subtitles,omitempty"`
	OutputSuffix string        `json:"outputSuffix,omitempty"`
	RawExtras    []string      `json:"rawExtras,omitempty"`
	Threading    ThreadingSpec `json:"threading,omitempty"`
}

type ContainerSpec struct {
	Format        string   `json:"format"` // mp4|mkv|webm|mov
	Faststart     bool     `json:"faststart,omitempty"`
	MovFlags      []string `json:"movflags,omitempty"`
	FragmentDurMs int      `json:"fragmentDurMs,omitempty"`
}

// UnmarshalJSON supports both the new object form {"format":"mp4"} and the
// legacy string form "mp4" so that presets written before the V2 refactor
// can still be read without a forced migration.
func (c *ContainerSpec) UnmarshalJSON(data []byte) error {
	// Try object form first.
	type plain ContainerSpec
	var obj plain
	if err := json.Unmarshal(data, &obj); err == nil {
		*c = ContainerSpec(obj)
		return nil
	}
	// Fallback: legacy string form.
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		c.Format = s
		return nil
	}
	return fmt.Errorf("container: expected object or string, got %s", string(data))
}

type InputSpec struct {
	SeekStart         string   `json:"seekStart,omitempty"` // -ss
	Duration          string   `json:"duration,omitempty"`  // -t
	ReadRateLimit     float64  `json:"readRateLimit,omitempty"`
	ProtocolWhitelist []string `json:"protocolWhitelist,omitempty"`
}

type HwaccelSpec struct {
	Type         string `json:"type"` // none|nvdec|vaapi|qsv|videotoolbox
	Device       string `json:"device,omitempty"`
	OutputFormat string `json:"outputFormat,omitempty"`
}

type RateMode string

const (
	RateModeCRF  RateMode = "crf"
	RateModeCBR  RateMode = "cbr"
	RateModeVBR  RateMode = "vbr"
	RateModeQP   RateMode = "qp"
	RateModeABR  RateMode = "abr"
	RateModeNone RateMode = "none"
)

type VideoRate struct {
	Mode    RateMode `json:"mode"`
	CRF     *int     `json:"crf,omitempty"`
	QP      *int     `json:"qp,omitempty"`
	Bitrate string   `json:"bitrate,omitempty"` // e.g. "5M", "5000k"
	MaxRate string   `json:"maxRate,omitempty"`
	BufSize string   `json:"bufSize,omitempty"`
	MinRate string   `json:"minRate,omitempty"`
}

type GopSpec struct {
	KeyintSec float64 `json:"keyintSec,omitempty"` // -g derived as keyintSec*fps
	KeyintMin int     `json:"keyintMin,omitempty"`
	SceneCut  *bool   `json:"sceneCut,omitempty"`
	BFrames   *int    `json:"bFrames,omitempty"`
	RefFrames *int    `json:"refFrames,omitempty"`
}

type ColorSpec struct {
	Range     string `json:"range,omitempty"`
	Primaries string `json:"primaries,omitempty"`
	Transfer  string `json:"transfer,omitempty"`
	Matrix    string `json:"matrix,omitempty"`
}

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type VideoSpec struct {
	Disabled    bool        `json:"disabled,omitempty"`
	Codec       string      `json:"codec"`
	Preset      string      `json:"preset,omitempty"`
	Tune        string      `json:"tune,omitempty"`
	Profile     string      `json:"profile,omitempty"`
	Level       string      `json:"level,omitempty"`
	PixelFormat string      `json:"pixelFormat,omitempty"`
	Rate        VideoRate   `json:"rate"`
	GOP         GopSpec     `json:"gop,omitempty"`
	TwoPass     bool        `json:"twoPass,omitempty"`
	Resolution  *Resolution `json:"resolution,omitempty"`
	FrameRate   string      `json:"frameRate,omitempty"`
	Color       *ColorSpec  `json:"color,omitempty"`
	CodecExtra  []KV        `json:"codecExtra,omitempty"`
}

type AudioSpec struct {
	Disabled   bool   `json:"disabled,omitempty"`
	Codec      string `json:"codec,omitempty"`
	Bitrate    string `json:"bitrate,omitempty"`
	SampleRate int    `json:"sampleRate,omitempty"`
	Channels   int    `json:"channels,omitempty"`
	Profile    string `json:"profile,omitempty"`
	VBRQuality *int   `json:"vbrQuality,omitempty"`
}

type FilterStep struct {
	Name    string            `json:"name"`
	Args    map[string]string `json:"args,omitempty"`
	Enabled *bool             `json:"enabled,omitempty"` // nil == enabled
}

func (s FilterStep) IsEnabled() bool {
	if s.Enabled == nil {
		return true
	}
	return *s.Enabled
}

type FilterChain struct {
	Video []FilterStep `json:"video,omitempty"`
	Audio []FilterStep `json:"audio,omitempty"`
}

type MappingSpec struct {
	Video    string   `json:"video,omitempty"` // "auto"|"all"|"none"|"0:v:0"...
	Audio    string   `json:"audio,omitempty"`
	Subtitle string   `json:"subtitle,omitempty"`
	Streams  []string `json:"streams,omitempty"` // explicit -map tokens
}

type SubtitleMode string

const (
	SubtitleCopy  SubtitleMode = "copy"
	SubtitleBurn  SubtitleMode = "burn"
	SubtitleStrip SubtitleMode = "strip"
)

type SubtitleSpec struct {
	Mode      SubtitleMode `json:"mode,omitempty"`
	BurnTrack int          `json:"burnTrack,omitempty"`
}

type ThreadingSpec struct {
	Threads       int `json:"threads,omitempty"`
	FilterThreads int `json:"filterThreads,omitempty"`
}

// UnmarshalSpec parses a raw spec jsonb document into a PresetSpec.
func UnmarshalSpec(raw []byte) (PresetSpec, error) {
	var s PresetSpec
	err := json.Unmarshal(raw, &s)
	return s, err
}
