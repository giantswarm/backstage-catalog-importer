package component

import "testing"

func TestIsValidLabelValue(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"10.3.0", true},
		{"2.3.0", true},
		{"0.15.0", true},
		{"a", true},
		{"1.0.0-rc.1", true},
		{"snake_case.v1", true},
		{"", false},
		{"dev:abc123", false},
		{"-leading", false},
		{"trailing.", false},
		{"has space", false},
		{"v/slash", false},
		{"0123456789012345678901234567890123456789012345678901234567890123", false}, // 64 chars
		{"012345678901234567890123456789012345678901234567890123456789012", true},   // 63 chars
	}

	for _, tt := range tests {
		if got := IsValidLabelValue(tt.value); got != tt.want {
			t.Errorf("IsValidLabelValue(%q) = %v, want %v", tt.value, got, tt.want)
		}
	}
}
