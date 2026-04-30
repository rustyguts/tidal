package workflows

import (
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// TriggerContext is what the watcher hands to the executor; its Vars() feed
// the template renderer.
type TriggerContext struct {
	Path     string
	Dir      string
	Filename string
	Stem     string
	Ext      string
	Date     time.Time
}

// NewTriggerContext fills in derived fields from a full source path.
func NewTriggerContext(fullPath string, t time.Time) TriggerContext {
	dir := filepath.Dir(fullPath)
	name := filepath.Base(fullPath)
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	return TriggerContext{
		Path:     fullPath,
		Dir:      dir,
		Filename: name,
		Stem:     stem,
		Ext:      ext,
		Date:     t,
	}
}

// Vars returns the `{{var}}` substitution map.
func (tc TriggerContext) Vars() map[string]string {
	return map[string]string{
		"path":     tc.Path,
		"dir":      tc.Dir,
		"filename": tc.Filename,
		"stem":     tc.Stem,
		"ext":      tc.Ext,
		"date":     tc.Date.Format("2006-01-02"),
	}
}

// `{{ key }}` with optional whitespace around the key.
var tmplPattern = regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`)

// Render walks the template and substitutes every `{{key}}` from vars.
// Unknown keys are left literal so users can spot a typo in stored job rows.
// No shell interpolation, no expression evaluation — keep the surface small.
func Render(tmpl string, vars map[string]string) string {
	return tmplPattern.ReplaceAllStringFunc(tmpl, func(match string) string {
		groups := tmplPattern.FindStringSubmatch(match)
		if len(groups) != 2 {
			return match
		}
		v, ok := vars[groups[1]]
		if !ok {
			return match
		}
		return v
	})
}
