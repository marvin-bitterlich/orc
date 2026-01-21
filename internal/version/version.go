package version

import (
	"fmt"
	"runtime"
)

// These variables are set at build time via ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

// String returns the full version string
func String() string {
	return fmt.Sprintf("orc %s (commit: %s, built: %s, %s/%s)",
		Version, shortCommit(), BuildTime, runtime.GOOS, runtime.GOARCH)
}

// Short returns just the version number
func Short() string {
	return Version
}

// Info returns structured version information
func Info() map[string]string {
	return map[string]string{
		"version":   Version,
		"commit":    Commit,
		"buildTime": BuildTime,
		"go":        runtime.Version(),
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
	}
}

func shortCommit() string {
	if len(Commit) > 7 {
		return Commit[:7]
	}
	return Commit
}
