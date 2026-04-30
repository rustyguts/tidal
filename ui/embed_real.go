//go:build embed

package ui

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

func init() {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	Dist = sub
	Available = true
}
