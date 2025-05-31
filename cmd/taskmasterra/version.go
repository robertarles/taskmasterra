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
		
	}
}

func getVersionString() string {
	return fmt.Sprintf("taskmasterra %s", Version)
} 