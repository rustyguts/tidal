// Package catalog is the single source of truth for the ffmpeg flag/codec
// surface that Tidal's preset system supports. The Catalog drives both Go-side
// validation and the JSON Schema served to the UI editor.
package catalog

type Catalog struct {
	Containers     []Container     `json:"containers"`
	VideoCodecs    []VideoCodec    `json:"videoCodecs"`
	AudioCodecs    []AudioCodec    `json:"audioCodecs"`
	PixelFormats   []string        `json:"pixelFormats"`
	Hwaccels       []Hwaccel       `json:"hwaccels"`
	Filters        []Filter        `json:"filters"`
	RawExtrasAllow []string        `json:"rawExtrasAllow"`
	RawExtrasDeny  []string        `json:"rawExtrasDeny"`
}

type Container struct {
	Format    string   `json:"format"`              // mp4|mkv|webm|mov
	Faststart bool     `json:"faststart,omitempty"` // -movflags +faststart applies
	MovFlags  []string `json:"movflags,omitempty"`  // valid -movflags tokens
	MIME      string   `json:"mime"`
}

type CodecFamily string

const (
	FamilyCPU      CodecFamily = "cpu"
	FamilyNVENC    CodecFamily = "nvenc"
	FamilyVAAPI    CodecFamily = "vaapi"
	FamilyQSV      CodecFamily = "qsv"
	FamilyPassthru CodecFamily = "passthrough"
)

type RateMode string

const (
	RateCRF  RateMode = "crf"
	RateCBR  RateMode = "cbr"
	RateVBR  RateMode = "vbr"
	RateQP   RateMode = "qp"
	RateABR  RateMode = "abr"
	RateNone RateMode = "none"
)

type IntRange struct {
	Min     int `json:"min"`
	Max     int `json:"max"`
	Default int `json:"default"`
}

type VideoCodec struct {
	Name            string      `json:"name"` // libx264, h264_nvenc, copy ...
	Family          CodecFamily `json:"family"`
	DisplayName     string      `json:"displayName"`
	Description     string      `json:"description,omitempty"`
	Presets         []string    `json:"presets,omitempty"` // -preset values
	Tunes           []string    `json:"tunes,omitempty"`
	Profiles        []string    `json:"profiles,omitempty"`
	Levels          []string    `json:"levels,omitempty"`
	PixelFormats    []string    `json:"pixelFormats,omitempty"`
	CRFRange        *IntRange   `json:"crfRange,omitempty"`
	QPRange         *IntRange   `json:"qpRange,omitempty"`
	RateModes       []RateMode  `json:"rateModes"`
	AllowsTwoPass   bool        `json:"allowsTwoPass"`
	ParamFlag       string      `json:"paramFlag,omitempty"` // -x264-params, -svtav1-params...
}

type AudioCodec struct {
	Name        string    `json:"name"` // aac, libopus, copy ...
	DisplayName string    `json:"displayName"`
	Description string    `json:"description,omitempty"`
	SampleRates []int     `json:"sampleRates,omitempty"`
	Channels    []int     `json:"channels,omitempty"`
	BitrateKbps *IntRange `json:"bitrateKbps,omitempty"`
	AllowsVBR   bool      `json:"allowsVBR"`
	VBRQuality  *IntRange `json:"vbrQuality,omitempty"`
	Profiles    []string  `json:"profiles,omitempty"`
}

type Hwaccel struct {
	Type        string   `json:"type"` // none|nvdec|vaapi|qsv|videotoolbox
	DisplayName string   `json:"displayName"`
	Description string   `json:"description,omitempty"`
	CodecNames  []string `json:"codecNames"`           // codecs that pair with this hwaccel
	OutputFmts  []string `json:"outputFormats"`        // -hwaccel_output_format values
	DefaultDev  string   `json:"defaultDevice,omitempty"`
}

type FilterScope string

const (
	FilterVideo FilterScope = "video"
	FilterAudio FilterScope = "audio"
)

type FilterArg struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"` // string|int|float|bool|enum
	Required    bool     `json:"required,omitempty"`
	Default     string   `json:"default,omitempty"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Min         *float64 `json:"min,omitempty"`
	Max         *float64 `json:"max,omitempty"`
}

type Filter struct {
	Name        string      `json:"name"`
	Scope       FilterScope `json:"scope"`
	DisplayName string      `json:"displayName"`
	Description string      `json:"description,omitempty"`
	Args        []FilterArg `json:"args,omitempty"`
}

// Default returns Tidal's curated catalog. Stable for the lifetime of the
// process — callers may share by pointer.
func Default() *Catalog {
	return &Catalog{
		Containers:     defaultContainers(),
		VideoCodecs:    defaultVideoCodecs(),
		AudioCodecs:    defaultAudioCodecs(),
		PixelFormats:   defaultPixelFormats(),
		Hwaccels:       defaultHwaccels(),
		Filters:        defaultFilters(),
		RawExtrasAllow: rawExtrasAllow(),
		RawExtrasDeny:  rawExtrasDeny(),
	}
}

func (c *Catalog) Container(format string) (Container, bool) {
	for _, x := range c.Containers {
		if x.Format == format {
			return x, true
		}
	}
	return Container{}, false
}

func (c *Catalog) VideoCodec(name string) (VideoCodec, bool) {
	for _, x := range c.VideoCodecs {
		if x.Name == name {
			return x, true
		}
	}
	return VideoCodec{}, false
}

func (c *Catalog) AudioCodec(name string) (AudioCodec, bool) {
	for _, x := range c.AudioCodecs {
		if x.Name == name {
			return x, true
		}
	}
	return AudioCodec{}, false
}

func (c *Catalog) Hwaccel(typ string) (Hwaccel, bool) {
	for _, x := range c.Hwaccels {
		if x.Type == typ {
			return x, true
		}
	}
	return Hwaccel{}, false
}

func (c *Catalog) Filter(name string) (Filter, bool) {
	for _, x := range c.Filters {
		if x.Name == name {
			return x, true
		}
	}
	return Filter{}, false
}

func defaultContainers() []Container {
	return []Container{
		{Format: "mp4", Faststart: true, MovFlags: []string{"+faststart", "+frag_keyframe", "+empty_moov", "+separate_moof"}, MIME: "video/mp4"},
		{Format: "mkv", MIME: "video/x-matroska"},
		{Format: "webm", MIME: "video/webm"},
		{Format: "mov", Faststart: true, MovFlags: []string{"+faststart"}, MIME: "video/quicktime"},
	}
}

func defaultPixelFormats() []string {
	return []string{
		"yuv420p", "yuv420p10le", "yuv422p", "yuv422p10le", "yuv444p", "yuv444p10le",
		"nv12", "nv21", "p010le", "p016le", "rgb24", "rgba", "gray",
	}
}
