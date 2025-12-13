package version

import (
	"fmt"
	"runtime"
)

// These variables are set during build time via ldflags
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
	BuiltBy = "unknown"
)

// Info returns formatted version information
func Info() string {
	return fmt.Sprintf("EndpointBOM %s\nCommit: %s\nBuilt: %s\nBuilt by: %s\nGo: %s\nOS/Arch: %s/%s",
		Version,
		Commit,
		Date,
		BuiltBy,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)
}

// Short returns a short version string
func Short() string {
	return Version
}

