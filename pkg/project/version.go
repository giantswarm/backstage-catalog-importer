package project

import (
	"fmt"
	"runtime/debug"
)

var (
	version        string
	gitSHA         string
	buildTimestamp string
)

// Version returns the release version and build metadata.
func Version() string {
	buildInfo, _ := debug.ReadBuildInfo()
	return formatVersion(buildInfo)
}

func formatVersion(buildInfo *debug.BuildInfo) string {
	releaseVersion, commit, date := version, gitSHA, buildTimestamp
	if buildInfo != nil {
		if releaseVersion == "" && buildInfo.Main.Version != "(devel)" {
			releaseVersion = buildInfo.Main.Version
		}
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				if commit == "" {
					commit = setting.Value
				}
			case "vcs.time":
				if date == "" {
					date = setting.Value
				}
			}
		}
	}
	if releaseVersion == "" {
		releaseVersion = "0.0.0-dev"
	}
	if commit == "" {
		commit = "unknown"
	}
	if date == "" {
		date = "unknown"
	}
	return fmt.Sprintf("%s (commit: %s, date: %s)", releaseVersion, commit, date)
}
