package catalog

func defaultAudioCodecs() []AudioCodec {
	commonRates := []int{8000, 11025, 16000, 22050, 32000, 44100, 48000, 88200, 96000}
	return []AudioCodec{
		{
			Name:        "aac",
			DisplayName: "AAC",
			Description: "Advanced Audio Coding — broadest compatibility for mp4/mov.",
			SampleRates: commonRates,
			Channels:    []int{1, 2, 6, 8},
			BitrateKbps: &IntRange{Min: 8, Max: 512, Default: 192},
			AllowsVBR:   true,
			VBRQuality:  &IntRange{Min: 1, Max: 5, Default: 4},
			Profiles:    []string{"aac_low", "aac_he", "aac_he_v2", "aac_ld", "aac_eld"},
		},
		{
			Name:        "libopus",
			DisplayName: "Opus",
			Description: "Modern, royalty-free; excellent at low bitrates. WebM-native.",
			SampleRates: []int{8000, 12000, 16000, 24000, 48000},
			Channels:    []int{1, 2, 6, 8},
			BitrateKbps: &IntRange{Min: 6, Max: 510, Default: 128},
			AllowsVBR:   true,
			VBRQuality:  &IntRange{Min: 0, Max: 10, Default: 5},
		},
		{
			Name:        "libmp3lame",
			DisplayName: "MP3",
			SampleRates: []int{8000, 11025, 12000, 16000, 22050, 24000, 32000, 44100, 48000},
			Channels:    []int{1, 2},
			BitrateKbps: &IntRange{Min: 8, Max: 320, Default: 192},
			AllowsVBR:   true,
			VBRQuality:  &IntRange{Min: 0, Max: 9, Default: 2},
		},
		{
			Name:        "flac",
			DisplayName: "FLAC (lossless)",
			SampleRates: commonRates,
			Channels:    []int{1, 2, 6, 8},
		},
		{
			Name:        "ac3",
			DisplayName: "AC-3 (Dolby Digital)",
			SampleRates: []int{32000, 44100, 48000},
			Channels:    []int{1, 2, 6},
			BitrateKbps: &IntRange{Min: 32, Max: 640, Default: 384},
		},
		{
			Name:        "eac3",
			DisplayName: "E-AC-3 (Dolby Digital Plus)",
			SampleRates: []int{32000, 44100, 48000},
			Channels:    []int{1, 2, 6, 8},
			BitrateKbps: &IntRange{Min: 32, Max: 6144, Default: 640},
		},
		{
			Name:        "copy",
			DisplayName: "Copy (no re-encode)",
			Description: "Stream copy — passthrough.",
		},
	}
}
