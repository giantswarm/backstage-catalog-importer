package component

import (
	"reflect"
	"testing"
)

func TestGenerateDeploymentNames(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "name with -app suffix",
			input:    "alertmanager-to-github-app",
			expected: []string{"alertmanager-to-github", "alertmanager-to-github-app"},
		},
		{
			name:     "name without -app suffix",
			input:    "alloy",
			expected: []string{"alloy", "alloy-app"},
		},
		{
			name:     "name with multiple hyphens and -app suffix",
			input:    "my-awesome-service-app",
			expected: []string{"my-awesome-service", "my-awesome-service-app"},
		},
		{
			name:     "simple name",
			input:    "nginx",
			expected: []string{"nginx", "nginx-app"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateDeploymentNames(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GenerateDeploymentNames(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

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
