package catalog

// rawExtrasAllow is the regex allowlist for tokens permitted in PresetSpecV2.RawExtras.
// Order does not matter — first match wins after deny is checked.
func rawExtrasAllow() []string {
	return []string{
		`^-(map|map_metadata|map_chapters)$`,
		`^-metadata(:s:[avs]:\d+)?$`,
		`^-disposition:[avs]:\d+$`,
		`^-(force_key_frames|g|keyint_min|sc_threshold|bf|refs|qmin|qmax)$`,
		`^-(x264|x265|svtav1|aom|vpx)-params$`,
		`^-(profile:v|level|tier)$`,
		`^-(profile:a|aq|q:a|vbr)$`,
		`^-(filter_complex|lavfi)$`,
		`^-shortest$`,
		`^-fflags$`,
		`^-avoid_negative_ts$`,
		`^-stats_period$`,
		`^-(rc|rc-lookahead|spatial-aq|temporal-aq|cq|look_ahead)$`,
		`^-(color_primaries|color_trc|colorspace|color_range)$`,
		`^-(application|frame_duration|packet_loss|cutoff)$`,
		`^-(vn|an|sn|dn)$`,
		`^-tag:[avsd]$`,
	}
}

// rawExtrasDeny is checked first; matches reject the token outright. Values are
// also screened: any value containing shell metacharacters is rejected.
func rawExtrasDeny() []string {
	return []string{
		`^-i$`,
		`^-y$`,
		`^-n$`,
		`^-f$`,
		`^-progress$`,
		`^-listen$`,
		`^-protocol_whitelist$`,
		`^-loglevel$`,
		`^-hide_banner$`,
		`^-pass$`,
		`^-passlogfile$`,
		`^-c:[vasx]$`, // codecs come from structured Video/Audio sections
		`^-vcodec$`,
		`^-acodec$`,
		`^-scodec$`,
		`^-(b:v|b:a|maxrate|minrate|bufsize|crf|qp)$`, // rate flags from structured Video.Rate
		`^-r$`, // framerate from structured Video.FrameRate
		`^-(ss|to|t|itsoffset)$`, // input timing from Input section
		`^-(hwaccel|hwaccel_device|hwaccel_output_format|vaapi_device|qsv_device|init_hw_device|filter_hw_device)$`,
		`^-(vf|af)$`, // filters from Filters section
		`^-(threads|filter_threads|filter_complex_threads)$`, // from Threading section
		`^-movflags$`, // from Container section
	}
}

// ValueMetacharacterPattern is applied to every value token (the arg after a
// flag) — match means reject. Also enforces an upper bound on length.
const ValueMaxLen = 256

var ValueDeny = []string{
	`[;|&]`,
	"`",
	`\$\(`,
	`\$\{`,
	`>`,
	`<\(`,
}
