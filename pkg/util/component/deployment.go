// Package component provides utility functions for component metadata handling.
package component

// IsChartDeployable returns true if a Helm chart with the given type is deployable.
// A chart is deployable if its type is "application" or empty (default is application).
// Library charts (type "library") are not deployable as they only provide reusable templates.
func IsChartDeployable(chartType string) bool {
	return chartType == "application" || chartType == ""
}
