package component

import (
	"testing"
)

func TestIsChartDeployable(t *testing.T) {
	tests := []struct {
		name      string
		chartType string
		expected  bool
	}{
		{
			name:      "application chart",
			chartType: "application",
			expected:  true,
		},
		{
			name:      "empty chart type (default is application)",
			chartType: "",
			expected:  true,
		},
		{
			name:      "library chart",
			chartType: "library",
			expected:  false,
		},
		{
			name:      "unknown chart type",
			chartType: "unknown",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsChartDeployable(tt.chartType)
			if result != tt.expected {
				t.Errorf("IsChartDeployable(%q) = %v, want %v", tt.chartType, result, tt.expected)
			}
		})
	}
}
