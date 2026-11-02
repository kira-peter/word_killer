package ui

import (
	"os"
	"strings"
)

// Version is the game version shown on the about page
const Version = "2.0.0"

// BuildTime can be overridden at link time, e.g.
//
//	make build BUILD_TIME=2026-11-02T17:20
var BuildTime = ""

// buildTimeLabel returns the build time for display. When BuildTime was not
// injected it falls back to the executable's modification time, which is the
// build time for a locally compiled binary.
func buildTimeLabel() string {
	if BuildTime != "" {
		// Link flags cannot carry spaces, so the value uses "2006-01-02T15:04"
		return strings.Replace(BuildTime, "T", " ", 1)
	}

	exe, err := os.Executable()
	if err != nil {
		return "unknown"
	}
	info, err := os.Stat(exe)
	if err != nil {
		return "unknown"
	}
	return info.ModTime().Format("2006-01-02 15:04")
}
