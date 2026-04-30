package version

// Build-time variables stamped via -ldflags.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

type Info struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

func Get() Info {
	return Info{Version: Version, Commit: Commit, Date: Date}
}
