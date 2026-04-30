// Package ui exposes the built Vue SPA as an io/fs.FS. The default build (no
// `embed` build tag) skips the embed so backend-only `go build` is fast and
// doesn't require ui/dist to exist. Production builds use `-tags=embed`.
package ui

import "io/fs"

// Dist holds the SPA filesystem when present. Always check Available first.
var Dist fs.FS

// Available reports whether the SPA is embedded into the binary.
var Available = false
