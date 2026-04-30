package catalog

func defaultFilters() []Filter {
	mn := func(v float64) *float64 { return &v }

	return []Filter{
		// Video
		{
			Name: "scale", Scope: FilterVideo, DisplayName: "Scale (resize)",
			Description: "Resize the video. Use -1 or -2 to preserve aspect ratio.",
			Args: []FilterArg{
				{Name: "w", Type: "string", Required: true, Description: "Width (px) or expression like iw/2"},
				{Name: "h", Type: "string", Required: true, Description: "Height (px) or expression like -2"},
				{Name: "flags", Type: "enum", Enum: []string{"fast_bilinear", "bilinear", "bicubic", "lanczos", "spline"}, Default: "bicubic"},
				{Name: "force_original_aspect_ratio", Type: "enum", Enum: []string{"disable", "decrease", "increase"}},
			},
		},
		{
			Name: "crop", Scope: FilterVideo, DisplayName: "Crop",
			Args: []FilterArg{
				{Name: "w", Type: "string", Required: true},
				{Name: "h", Type: "string", Required: true},
				{Name: "x", Type: "string", Default: "(in_w-out_w)/2"},
				{Name: "y", Type: "string", Default: "(in_h-out_h)/2"},
			},
		},
		{
			Name: "pad", Scope: FilterVideo, DisplayName: "Pad",
			Args: []FilterArg{
				{Name: "width", Type: "string", Required: true},
				{Name: "height", Type: "string", Required: true},
				{Name: "x", Type: "string", Default: "0"},
				{Name: "y", Type: "string", Default: "0"},
				{Name: "color", Type: "string", Default: "black"},
			},
		},
		{
			Name: "fps", Scope: FilterVideo, DisplayName: "Frame rate",
			Args: []FilterArg{
				{Name: "fps", Type: "string", Required: true, Description: "e.g. 24, 30, 60000/1001"},
				{Name: "round", Type: "enum", Enum: []string{"zero", "inf", "down", "up", "near"}, Default: "near"},
			},
		},
		{Name: "setsar", Scope: FilterVideo, DisplayName: "Set sample aspect ratio",
			Args: []FilterArg{{Name: "ratio", Type: "string", Required: true, Default: "1/1"}}},
		{Name: "setdar", Scope: FilterVideo, DisplayName: "Set display aspect ratio",
			Args: []FilterArg{{Name: "ratio", Type: "string", Required: true, Default: "16/9"}}},
		{Name: "format", Scope: FilterVideo, DisplayName: "Pixel format",
			Args: []FilterArg{{Name: "pix_fmts", Type: "string", Required: true, Default: "yuv420p"}}},
		{Name: "transpose", Scope: FilterVideo, DisplayName: "Transpose / rotate",
			Args: []FilterArg{{Name: "dir", Type: "enum", Enum: []string{"0", "1", "2", "3"}, Required: true, Description: "0=ccw+vflip, 1=cw, 2=ccw, 3=cw+vflip"}}},
		{Name: "hflip", Scope: FilterVideo, DisplayName: "Horizontal flip"},
		{Name: "vflip", Scope: FilterVideo, DisplayName: "Vertical flip"},
		{
			Name: "unsharp", Scope: FilterVideo, DisplayName: "Unsharp mask",
			Args: []FilterArg{
				{Name: "lx", Type: "int", Default: "5"},
				{Name: "ly", Type: "int", Default: "5"},
				{Name: "la", Type: "float", Default: "1.0"},
				{Name: "cx", Type: "int", Default: "5"},
				{Name: "cy", Type: "int", Default: "5"},
				{Name: "ca", Type: "float", Default: "0.0"},
			},
		},
		{
			Name: "eq", Scope: FilterVideo, DisplayName: "Color adjust (eq)",
			Args: []FilterArg{
				{Name: "contrast", Type: "float", Default: "1.0"},
				{Name: "brightness", Type: "float", Default: "0.0"},
				{Name: "saturation", Type: "float", Default: "1.0"},
				{Name: "gamma", Type: "float", Default: "1.0"},
			},
		},
		{
			Name: "yadif", Scope: FilterVideo, DisplayName: "Deinterlace (yadif)",
			Args: []FilterArg{
				{Name: "mode", Type: "enum", Enum: []string{"0", "1", "2", "3"}, Default: "0"},
				{Name: "parity", Type: "enum", Enum: []string{"-1", "0", "1"}, Default: "-1"},
				{Name: "deint", Type: "enum", Enum: []string{"0", "1"}, Default: "0"},
			},
		},
		{
			Name: "bwdif", Scope: FilterVideo, DisplayName: "Deinterlace (bwdif, higher quality)",
			Args: []FilterArg{
				{Name: "mode", Type: "enum", Enum: []string{"0", "1"}, Default: "0"},
				{Name: "parity", Type: "enum", Enum: []string{"-1", "0", "1"}, Default: "-1"},
			},
		},
		{
			Name: "subtitles", Scope: FilterVideo, DisplayName: "Burn subtitles",
			Args: []FilterArg{
				{Name: "filename", Type: "string", Required: true, Description: "Path to .srt/.ass — must be in an allowed media root"},
				{Name: "stream_index", Type: "int"},
				{Name: "force_style", Type: "string", Description: "ASS style overrides"},
			},
		},
		{
			Name: "drawtext", Scope: FilterVideo, DisplayName: "Draw text",
			Args: []FilterArg{
				{Name: "text", Type: "string", Required: true},
				{Name: "fontfile", Type: "string"},
				{Name: "fontsize", Type: "int", Default: "24"},
				{Name: "fontcolor", Type: "string", Default: "white"},
				{Name: "x", Type: "string", Default: "10"},
				{Name: "y", Type: "string", Default: "10"},
				{Name: "box", Type: "bool"},
				{Name: "boxcolor", Type: "string"},
			},
		},
		{
			Name: "overlay", Scope: FilterVideo, DisplayName: "Overlay (e.g. logo)",
			Args: []FilterArg{
				{Name: "x", Type: "string", Default: "0"},
				{Name: "y", Type: "string", Default: "0"},
				{Name: "format", Type: "enum", Enum: []string{"yuv420", "yuv422", "yuv444", "rgb", "auto"}, Default: "auto"},
			},
		},
		{Name: "colorspace", Scope: FilterVideo, DisplayName: "Convert color space",
			Args: []FilterArg{{Name: "all", Type: "string"}, {Name: "iall", Type: "string"}, {Name: "trc", Type: "string"}, {Name: "primaries", Type: "string"}, {Name: "space", Type: "string"}, {Name: "range", Type: "string"}}},
		{
			Name: "tonemap", Scope: FilterVideo, DisplayName: "HDR → SDR tone-mapping",
			Args: []FilterArg{
				{Name: "tonemap", Type: "enum", Enum: []string{"none", "linear", "gamma", "clip", "reinhard", "hable", "mobius"}, Default: "hable"},
				{Name: "param", Type: "float"},
				{Name: "desat", Type: "float", Default: "2"},
			},
		},
		{Name: "zscale", Scope: FilterVideo, DisplayName: "zscale (high-quality color/range conversion)",
			Args: []FilterArg{
				{Name: "w", Type: "string"},
				{Name: "h", Type: "string"},
				{Name: "matrix", Type: "string"},
				{Name: "transfer", Type: "string"},
				{Name: "primaries", Type: "string"},
				{Name: "range", Type: "string"},
			},
		},
		{Name: "hwupload", Scope: FilterVideo, DisplayName: "Upload frame to GPU"},
		{Name: "hwdownload", Scope: FilterVideo, DisplayName: "Download frame from GPU"},

		// Audio
		{
			Name: "loudnorm", Scope: FilterAudio, DisplayName: "EBU R128 loudness normalize",
			Args: []FilterArg{
				{Name: "I", Type: "float", Default: "-16", Min: mn(-70), Max: mn(-5), Description: "Integrated LUFS target"},
				{Name: "TP", Type: "float", Default: "-1.5", Description: "True peak target dBTP"},
				{Name: "LRA", Type: "float", Default: "11"},
				{Name: "linear", Type: "bool", Default: "true"},
			},
		},
		{
			Name: "dynaudnorm", Scope: FilterAudio, DisplayName: "Dynamic audio normalize",
			Args: []FilterArg{
				{Name: "f", Type: "int", Default: "500", Description: "Frame length (ms)"},
				{Name: "g", Type: "int", Default: "31"},
				{Name: "p", Type: "float", Default: "0.95"},
			},
		},
		{
			Name: "aresample", Scope: FilterAudio, DisplayName: "Resample audio",
			Args: []FilterArg{
				{Name: "sample_rate", Type: "int", Required: true},
				{Name: "resampler", Type: "enum", Enum: []string{"swr", "soxr"}, Default: "swr"},
			},
		},
		{Name: "asetrate", Scope: FilterAudio, DisplayName: "Pitch-shift via samplerate change",
			Args: []FilterArg{{Name: "sample_rate", Type: "int", Required: true}}},
		{Name: "silenceremove", Scope: FilterAudio, DisplayName: "Trim silence",
			Args: []FilterArg{
				{Name: "start_periods", Type: "int", Default: "1"},
				{Name: "start_threshold", Type: "string", Default: "-50dB"},
				{Name: "stop_periods", Type: "int", Default: "0"},
			},
		},
		{
			Name: "volume", Scope: FilterAudio, DisplayName: "Volume",
			Args: []FilterArg{{Name: "volume", Type: "string", Required: true, Default: "1.0", Description: "e.g. 1.5, 0.5, -3dB"}},
		},
		{
			Name: "pan", Scope: FilterAudio, DisplayName: "Channel mix / pan",
			Args: []FilterArg{{Name: "spec", Type: "string", Required: true, Description: "e.g. stereo|c0=c0|c1=c1"}},
		},
		{Name: "aformat", Scope: FilterAudio, DisplayName: "Force audio format",
			Args: []FilterArg{
				{Name: "sample_fmts", Type: "string"},
				{Name: "sample_rates", Type: "string"},
				{Name: "channel_layouts", Type: "string"},
			},
		},
		{
			Name: "acompressor", Scope: FilterAudio, DisplayName: "Audio compressor",
			Args: []FilterArg{
				{Name: "threshold", Type: "string", Default: "0.125"},
				{Name: "ratio", Type: "float", Default: "2"},
				{Name: "attack", Type: "float", Default: "20"},
				{Name: "release", Type: "float", Default: "250"},
				{Name: "makeup", Type: "float", Default: "1"},
			},
		},
		{
			Name: "highpass", Scope: FilterAudio, DisplayName: "High-pass filter",
			Args: []FilterArg{
				{Name: "frequency", Type: "int", Default: "70"},
				{Name: "poles", Type: "enum", Enum: []string{"1", "2"}, Default: "2"},
			},
		},
		{
			Name: "lowpass", Scope: FilterAudio, DisplayName: "Low-pass filter",
			Args: []FilterArg{
				{Name: "frequency", Type: "int", Default: "8000"},
				{Name: "poles", Type: "enum", Enum: []string{"1", "2"}, Default: "2"},
			},
		},
	}
}
