// Package component provides utility functions for component metadata handling.
package component

import (
	"fmt"
	"strings"
)

// GenerateDeploymentNames generates default deployment names from a component name.
// It returns two variants: one without the "-app" suffix and one with it.
//
// Examples:
//   - "alertmanager-to-github-app" -> ["alertmanager-to-github", "alertmanager-to-github-app"]
//   - "alloy" -> ["alloy", "alloy-app"]
func GenerateDeploymentNames(name string) []string {
	baseName := strings.TrimSuffix(name, "-app")
	nameWithAppSuffix := fmt.Sprintf("%s-app", baseName)
	return []string{
		baseName,
		nameWithAppSuffix,
	}
}

// IsChartDeployable returns true if a Helm chart with the given type is deployable.
// A chart is deployable if its type is "application" or empty (default is application).
// Library charts (type "library") are not deployable as they only provide reusable templates.
func IsChartDeployable(chartType string) bool {
	return chartType == "application" || chartType == ""
}
