package main

import (
	"fmt"
	"runtime/debug"
)

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		// Get version from module version if not set via ldflags
		if Version == "dev" && info.Main.Version != "(devel)" && info.Main.Version != "" {
			Version = info.Main.Version
		}
		
		// Always try to get commit and build time if not set via ldflags
		if Commit == "none" || BuildTime == "unknown" {
			for _, setting := range info.Settings {
				switch setting.Key {
				case "vcs.revision":
					if Commit == "none" {
						Commit = setting.Value[:7] // First 7 chars of commit hash
					}
				case "vcs.time":
					if BuildTime == "unknown" {
						BuildTime = setting.Value
					}
				}
			}
		}
	}
}

func getVersionString() string {
	return fmt.Sprintf("taskmasterra %s (commit: %s, built at: %s)", Version, Commit, BuildTime)
} 