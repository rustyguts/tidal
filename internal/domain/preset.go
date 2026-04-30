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
	Spec        PresetSpec `json:"spec"`
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
