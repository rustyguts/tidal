package builder

import "github.com/rustyguts/tidal/internal/domain"

// rawExtras appends user-supplied passthrough flags. Catalog allow/deny
// validation lives in the validator (internal/domain/preset_v2_validate.go),
// which runs at preset Create/Update time. The builder trusts its input.
func rawExtras(_ Context, spec domain.PresetSpec) ([]string, error) {
	if len(spec.RawExtras) == 0 {
		return nil, nil
	}
	out := make([]string, len(spec.RawExtras))
	copy(out, spec.RawExtras)
	return out, nil
}
