package component

import "testing"

func TestIsPublishableChartVersion(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		// Real releases, the only shape chart versions take in the catalog today.
		{"1.2.3", true},
		{"0.4.1", true},
		{"v1.6.0", true},
		{"0.0.1", true},

		// Unsubstituted placeholders, read from a repo's default branch.
		{"0.0.0-dev", false},
		{"1.6.0-dev", false},
		{"0.2.383-dev", false},
		{"0.0.0", false},
		{"", false},

		// Every spelling of the placeholder, all equally unsubstituted. These
		// carry no prerelease, so the shared IsReleaseVersion rule accepts them
		// and only the placeholder check rejects them.
		{"v0.0.0", false},
		{"0.0.0+abc1234", false},
		{"v0.0.0+abc1234", false},

		// Not a version at all.
		{"main", false},
	}

	for _, tt := range tests {
		if got := isPublishableChartVersion(tt.version); got != tt.want {
			t.Errorf("isPublishableChartVersion(%q) = %v, want %v", tt.version, got, tt.want)
		}
	}
}

func TestIsPublishableAppVersion(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		// Unsubstituted placeholders. These are the only values to drop.
		{"3.3.1-dev", false},
		{"0.0.0-dev", false},
		{"1.0.1-dev", false},
		{"0.0.0", false},
		{"v0.0.0", false},
		{"0.0.0+abc1234", false},
		{"", false},

		// Real upstream versions that a pure-semver rule would wrongly discard.
		// Every one of these is in the catalog today.
		{"v1.13.2", true},
		{"1.12.8-gs-6158079b4", true},
		{"v1.23.0-gs-c0faf6a76", true},
		{"0.17.0-gs.0", true},
		{"edge-25.12.3", true},
		{"main", true},
		{"master", true},
		{"4.9", true},
		{"gateway-api: v1.5.1", true},
	}

	for _, tt := range tests {
		if got := isPublishableAppVersion(tt.version); got != tt.want {
			t.Errorf("isPublishableAppVersion(%q) = %v, want %v", tt.version, got, tt.want)
		}
	}
}

func TestHasAny(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   bool
	}{
		{"nil", nil, false},
		{"all empty", []string{"", ""}, false},
		{"first slot filled", []string{"1.2.3", ""}, true},
		{"later slot filled", []string{"", "2.3.4"}, true},
	}

	for _, tt := range tests {
		if got := hasAny(tt.values); got != tt.want {
			t.Errorf("%s: hasAny(%v) = %v, want %v", tt.name, tt.values, got, tt.want)
		}
	}
}
