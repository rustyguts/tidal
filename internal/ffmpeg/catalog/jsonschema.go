package catalog

import "encoding/json"

// JSONSchema returns a draft-07 JSON Schema describing PresetSpec derived
// from the catalog. Cross-field constraints (codec ↔ hwaccel pairing,
// codec-specific preset enums) are intentionally not expressed here — those
// live in the Go validator. The schema is for input scaffolding only.
func JSONSchema(c *Catalog) json.RawMessage {
	containerFormats := make([]string, 0, len(c.Containers))
	for _, x := range c.Containers {
		containerFormats = append(containerFormats, x.Format)
	}
	videoCodecNames := make([]string, 0, len(c.VideoCodecs))
	for _, x := range c.VideoCodecs {
		videoCodecNames = append(videoCodecNames, x.Name)
	}
	audioCodecNames := make([]string, 0, len(c.AudioCodecs))
	for _, x := range c.AudioCodecs {
		audioCodecNames = append(audioCodecNames, x.Name)
	}
	hwaccelTypes := make([]string, 0, len(c.Hwaccels))
	for _, x := range c.Hwaccels {
		hwaccelTypes = append(hwaccelTypes, x.Type)
	}
	filterNames := make([]string, 0, len(c.Filters))
	for _, x := range c.Filters {
		filterNames = append(filterNames, x.Name)
	}

	schema := map[string]any{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"title":   "PresetSpec",
		"type":    "object",
		"required": []string{"container", "video", "audio"},
		"properties": map[string]any{
			"container": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"format":        map[string]any{"type": "string", "enum": containerFormats},
					"faststart":     map[string]any{"type": "boolean"},
					"movflags":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"fragmentDurMs": map[string]any{"type": "integer", "minimum": 0},
				},
				"required": []string{"format"},
			},
			"input": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"seekStart":         map[string]any{"type": "string"},
					"duration":          map[string]any{"type": "string"},
					"readRateLimit":     map[string]any{"type": "number", "minimum": 0},
					"protocolWhitelist": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				},
			},
			"hwaccel": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"type":         map[string]any{"type": "string", "enum": hwaccelTypes},
					"device":       map[string]any{"type": "string"},
					"outputFormat": map[string]any{"type": "string"},
				},
				"required": []string{"type"},
			},
			"video": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"disabled":    map[string]any{"type": "boolean"},
					"codec":       map[string]any{"type": "string", "enum": videoCodecNames},
					"preset":      map[string]any{"type": "string"},
					"tune":        map[string]any{"type": "string"},
					"profile":     map[string]any{"type": "string"},
					"level":       map[string]any{"type": "string"},
					"pixelFormat": map[string]any{"type": "string", "enum": c.PixelFormats},
					"rate": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"mode":    map[string]any{"type": "string", "enum": []string{"crf", "cbr", "vbr", "qp", "abr", "none"}},
							"crf":     map[string]any{"type": "integer", "minimum": 0, "maximum": 63},
							"qp":      map[string]any{"type": "integer", "minimum": 0, "maximum": 63},
							"bitrate": map[string]any{"type": "string"},
							"maxRate": map[string]any{"type": "string"},
							"bufSize": map[string]any{"type": "string"},
							"minRate": map[string]any{"type": "string"},
						},
					},
					"gop": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"keyintSec": map[string]any{"type": "number", "minimum": 0},
							"keyintMin": map[string]any{"type": "integer", "minimum": 0},
							"sceneCut":  map[string]any{"type": "boolean"},
							"bFrames":   map[string]any{"type": "integer", "minimum": 0, "maximum": 16},
							"refFrames": map[string]any{"type": "integer", "minimum": 0, "maximum": 16},
						},
					},
					"twoPass": map[string]any{"type": "boolean"},
					"resolution": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"width":  map[string]any{"type": "integer", "minimum": 2},
							"height": map[string]any{"type": "integer", "minimum": 2},
						},
					},
					"frameRate": map[string]any{"type": "string"},
					"color": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"range":     map[string]any{"type": "string"},
							"primaries": map[string]any{"type": "string"},
							"transfer":  map[string]any{"type": "string"},
							"matrix":    map[string]any{"type": "string"},
						},
					},
					"codecExtra": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "object", "properties": map[string]any{"key": map[string]any{"type": "string"}, "value": map[string]any{"type": "string"}}},
					},
				},
				"required": []string{"codec"},
			},
			"audio": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"disabled":   map[string]any{"type": "boolean"},
					"codec":      map[string]any{"type": "string", "enum": audioCodecNames},
					"bitrate":    map[string]any{"type": "string"},
					"sampleRate": map[string]any{"type": "integer", "minimum": 8000, "maximum": 192000},
					"channels":   map[string]any{"type": "integer", "minimum": 1, "maximum": 8},
					"profile":    map[string]any{"type": "string"},
					"vbrQuality": map[string]any{"type": "integer", "minimum": 0, "maximum": 10},
				},
				"required": []string{"codec"},
			},
			"filters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"video": filterChainSchema(filterNames),
					"audio": filterChainSchema(filterNames),
				},
			},
			"mapping": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"video":    map[string]any{"type": "string"},
					"audio":    map[string]any{"type": "string"},
					"subtitle": map[string]any{"type": "string"},
				},
			},
			"subtitles": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"mode":      map[string]any{"type": "string", "enum": []string{"copy", "burn", "strip"}},
					"burnTrack": map[string]any{"type": "integer", "minimum": 0},
				},
			},
			"outputSuffix": map[string]any{"type": "string"},
			"rawExtras":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"threading": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"threads":       map[string]any{"type": "integer", "minimum": 0},
					"filterThreads": map[string]any{"type": "integer", "minimum": 0},
				},
			},
		},
	}

	bytes, err := json.Marshal(schema)
	if err != nil {
		// Catalog → schema conversion is fully deterministic; a marshal error
		// here means a code bug, not a runtime input issue.
		panic("catalog: JSONSchema marshal failed: " + err.Error())
	}
	return bytes
}

func filterChainSchema(filterNames []string) map[string]any {
	return map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "object",
			"required": []string{"name"},
			"properties": map[string]any{
				"name": map[string]any{"type": "string", "enum": filterNames},
				"args": map[string]any{
					"type":                 "object",
					"additionalProperties": map[string]any{"type": "string"},
				},
				"enabled": map[string]any{"type": "boolean"},
			},
		},
	}
}
