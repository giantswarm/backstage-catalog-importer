package project

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatVersion(t *testing.T) {
	originalVersion, originalCommit, originalDate := version, gitSHA, buildTimestamp
	t.Cleanup(func() {
		version, gitSHA, buildTimestamp = originalVersion, originalCommit, originalDate
	})

	testCases := []struct {
		name           string
		version        string
		commit         string
		date           string
		buildInfo      *debug.BuildInfo
		expectedOutput string
	}{
		{
			name:           "missing build info",
			expectedOutput: "0.0.0-dev (commit: unknown, date: unknown)",
		},
		{
			name:           "development module",
			buildInfo:      &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			expectedOutput: "0.0.0-dev (commit: unknown, date: unknown)",
		},
		{
			name:           "empty module version",
			buildInfo:      &debug.BuildInfo{},
			expectedOutput: "0.0.0-dev (commit: unknown, date: unknown)",
		},
		{
			name: "release module with VCS metadata",
			buildInfo: &debug.BuildInfo{
				Main: debug.Module{Version: "v1.2.3"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abc123"},
					{Key: "vcs.time", Value: "2026-11-20T10:00:00Z"},
				},
			},
			expectedOutput: "v1.2.3 (commit: abc123, date: 2026-11-20T10:00:00Z)",
		},
		{
			name:    "ldflags override module and VCS metadata",
			version: "1.2.4",
			commit:  "def456",
			date:    "2026-11-20T11:00:00Z",
			buildInfo: &debug.BuildInfo{
				Main: debug.Module{Version: "v1.2.3"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abc123"},
					{Key: "vcs.time", Value: "2026-11-20T10:00:00Z"},
				},
			},
			expectedOutput: "1.2.4 (commit: def456, date: 2026-11-20T11:00:00Z)",
		},
		{
			name:    "partial ldflags retain VCS fallback",
			version: "1.2.4",
			buildInfo: &debug.BuildInfo{
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abc123"},
					{Key: "vcs.time", Value: "2026-11-20T10:00:00Z"},
				},
			},
			expectedOutput: "1.2.4 (commit: abc123, date: 2026-11-20T10:00:00Z)",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			version, gitSHA, buildTimestamp = testCase.version, testCase.commit, testCase.date
			require.Equal(t, testCase.expectedOutput, formatVersion(testCase.buildInfo))
		})
	}
}
