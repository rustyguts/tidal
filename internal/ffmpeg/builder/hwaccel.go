package builder

import (
	"fmt"
	"strings"

	"github.com/rustyguts/tidal/internal/domain"
)

// hwaccelInput emits pre-input hardware-acceleration flags. Cross-checks the
// hwaccel type against the chosen video codec family — mismatched pairings
// fail here with a structured error rather than at ffmpeg runtime.
func hwaccelInput(_ Context, spec domain.PresetSpecV2) ([]string, error) {
	if spec.Hwaccel == nil || spec.Hwaccel.Type == "" || spec.Hwaccel.Type == "none" {
		return nil, nil
	}
	hw := spec.Hwaccel
	codec := strings.ToLower(spec.Video.Codec)
	out := []string{}
	switch hw.Type {
	case "nvdec":
		if !strings.Contains(codec, "nvenc") {
			return nil, fmt.Errorf("hwaccel %q requires NVENC encoder, got %q", hw.Type, codec)
		}
		out = append(out, "-hwaccel", "cuda")
		fmtName := hw.OutputFormat
		if fmtName == "" {
			fmtName = "cuda"
		}
		out = append(out, "-hwaccel_output_format", fmtName)
		if hw.Device != "" {
			out = append(out, "-hwaccel_device", hw.Device)
		}
	case "vaapi":
		if !strings.Contains(codec, "vaapi") {
			return nil, fmt.Errorf("hwaccel %q requires VAAPI encoder, got %q", hw.Type, codec)
		}
		dev := hw.Device
		if dev == "" {
			dev = "/dev/dri/renderD128"
		}
		out = append(out, "-vaapi_device", dev, "-hwaccel", "vaapi")
		fmtName := hw.OutputFormat
		if fmtName == "" {
			fmtName = "vaapi"
		}
		out = append(out, "-hwaccel_output_format", fmtName)
	case "qsv":
		if !strings.Contains(codec, "qsv") {
			return nil, fmt.Errorf("hwaccel %q requires QSV encoder, got %q", hw.Type, codec)
		}
		dev := hw.Device
		if dev == "" {
			dev = "/dev/dri/renderD128"
		}
		out = append(out, "-qsv_device", dev, "-hwaccel", "qsv")
		fmtName := hw.OutputFormat
		if fmtName == "" {
			fmtName = "qsv"
		}
		out = append(out, "-hwaccel_output_format", fmtName)
	case "videotoolbox":
		out = append(out, "-hwaccel", "videotoolbox")
		if hw.OutputFormat != "" {
			out = append(out, "-hwaccel_output_format", hw.OutputFormat)
		}
	default:
		return nil, fmt.Errorf("unknown hwaccel type %q", hw.Type)
	}
	return out, nil
}
