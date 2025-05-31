package main

import (
	"fmt"
	"runtime/debug"
)

func init() {
	// If version info wasn't provided via ldflags, try to get it from build info
	if Version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok {
			// Get version from module version
			if info.Main.Version != "(devel)" && info.Main.Version != "" {
				Version = info.Main.Version
			}
			
			// Get commit from vcs information
			for _, setting := range info.Settings {
				switch setting.Key {
				case "vcs.revision":
					Commit = setting.Value[:7] // First 7 chars of commit hash
				case "vcs.time":
					BuildTime = setting.Value
				}
			}
		}
	}
}

func getVersionString() string {
	return fmt.Sprintf("taskmasterra %s (commit: %s, built at: %s)", Version, Commit, BuildTime)
} 