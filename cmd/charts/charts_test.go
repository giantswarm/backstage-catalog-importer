package charts

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/output/bscatalog/v1alpha1"
)

func TestCreateComponentFromOCIChart(t *testing.T) {
	// Fixed time for testing
	fixedTime, _ := time.Parse(time.RFC3339, "2023-10-15T10:30:00Z")

	tests := []struct {
		name                  string
		repo                  string
		tag                   string
		configMap             map[string]interface{}
		namespace             string
		componentType         string
		registryHostname      string
		wantName              string
		wantDescription       string
		wantTags              []string
		wantAnnotations       map[string]string
		wantLinks             []bscatalog.EntityLink
		wantVersion           string
		wantOwner             string
		wantGithubProjectSlug string
		wantCreatedTimeSet    bool
		wantErr               bool
	}{
		{
			name: "Chart with valid GitHub home URL",
			repo: "giantswarm/my-chart",
			tag:  "v1.0.0",
			configMap: map[string]interface{}{
				"home": "https://github.com/giantswarm/my-chart-app",
			},
			namespace:        "default",
			componentType:    "service",
			registryHostname: "registry.example.com",
			wantName:         "my-chart-app",
			wantDescription:  "OCI chart from giantswarm/my-chart",
			wantTags:         []string{"helmchart", "helmchart-deployable"},
			wantAnnotations: map[string]string{
				"application.giantswarm.io/audience": "all",
				"application.giantswarm.io/managed":  "false",
				"giantswarm.io/helmcharts":           "registry.example.com/giantswarm/my-chart",
				"giantswarm.io/helmchart-versions":   "v1.0.0",
			},
			wantLinks:             nil,
			wantVersion:           "v1.0.0",
			wantOwner:             "group:default/unspecified",
			wantGithubProjectSlug: "giantswarm/my-chart-app",
			wantCreatedTimeSet:    true,
			wantErr:               false,
		},
		{
			name: "Chart with full metadata and appVersion",
			repo: "giantswarm/advanced-chart",
			tag:  "v2.1.0",
			configMap: map[string]interface{}{
				"description": "Advanced Helm chart for testing",
				"version":     "2.1.0",
				"appVersion":  "1.5.3",
				"type":        "library",
				"created":     "2023-10-15T10:30:00Z",
				"home":        "https://github.com/giantswarm/advanced-chart-app",
			},
			namespace:        "giantswarm",
			componentType:    "library",
			registryHostname: "localhost:5000",
			wantName:         "advanced-chart-app",
			wantDescription:  "Advanced Helm chart for testing",
			wantTags:         []string{"helmchart"}, // library type - not deployable
			wantAnnotations: map[string]string{
				"application.giantswarm.io/audience":   "all",
				"application.giantswarm.io/managed":    "false",
				"giantswarm.io/helmcharts":             "localhost:5000/giantswarm/advanced-chart",
				"giantswarm.io/helmchart-versions":     "2.1.0",
				"giantswarm.io/helmchart-app-versions": "1.5.3",
			},
			wantLinks:             nil,
			wantVersion:           "2.1.0",
			wantOwner:             "group:giantswarm/unspecified",
			wantGithubProjectSlug: "giantswarm/advanced-chart-app",
			wantCreatedTimeSet:    true,
			wantErr:               false,
		},
		{
			name: "Chart with icon and team owner",
			repo: "giantswarm/chart-with-icon",
			tag:  "v1.0.0",
			configMap: map[string]interface{}{
				"description": "Chart with icon",
				"icon":        "https://example.com/icon.png",
				"home":        "https://github.com/giantswarm/chart-with-icon-app",
				"annotations": map[string]interface{}{
					"application.giantswarm.io/team": "honeybadger",
				},
			},
			namespace:        "default",
			componentType:    "service",
			registryHostname: "registry.example.com",
			wantName:         "chart-with-icon-app",
			wantDescription:  "Chart with icon",
			wantTags:         []string{"helmchart", "helmchart-deployable"},
			wantAnnotations: map[string]string{
				"application.giantswarm.io/audience": "all",
				"application.giantswarm.io/managed":  "false",
				"giantswarm.io/helmcharts":           "registry.example.com/giantswarm/chart-with-icon",
				"giantswarm.io/helmchart-versions":   "v1.0.0",
				"giantswarm.io/icon-url":             "https://example.com/icon.png",
			},
			wantLinks:             nil,
			wantVersion:           "v1.0.0",
			wantOwner:             "group:default/team-honeybadger",
			wantGithubProjectSlug: "giantswarm/chart-with-icon-app",
			wantCreatedTimeSet:    true,
			wantErr:               false,
		},
		{
			name: "Chart with team owner with team- prefix",
			repo: "giantswarm/atlas-chart",
			tag:  "v2.0.0",
			configMap: map[string]interface{}{
				"description": "Chart owned by team-atlas",
				"home":        "https://github.com/giantswarm/atlas-chart-app",
				"annotations": map[string]interface{}{
					"application.giantswarm.io/team": "team-atlas",
				},
			},
			namespace:        "default",
			componentType:    "service",
			registryHostname: "registry.example.com",
			wantName:         "atlas-chart-app",
			wantDescription:  "Chart owned by team-atlas",
			wantTags:         []string{"helmchart", "helmchart-deployable"},
			wantAnnotations: map[string]string{
				"application.giantswarm.io/audience": "all",
				"application.giantswarm.io/managed":  "false",
				"giantswarm.io/helmcharts":           "registry.example.com/giantswarm/atlas-chart",
				"giantswarm.io/helmchart-versions":   "v2.0.0",
			},
			wantLinks:             nil,
			wantVersion:           "v2.0.0",
			wantOwner:             "group:default/team-atlas",
			wantGithubProjectSlug: "giantswarm/atlas-chart-app",
			wantCreatedTimeSet:    true,
			wantErr:               false,
		},
		{
			name: "Chart with audience and managed annotations",
			repo: "giantswarm/managed-chart",
			tag:  "v1.5.0",
			configMap: map[string]interface{}{
				"description": "Managed Chart for GiantSwarm",
				"home":        "https://github.com/giantswarm/managed-chart-app",
				"annotations": map[string]interface{}{
					"application.giantswarm.io/audience": "giantswarm",
					"application.giantswarm.io/managed":  "true",
				},
			},
			namespace:        "custom",
			componentType:    "resource",
			registryHostname: "gsoci.azurecr.io",
			wantName:         "managed-chart-app",
			wantDescription:  "Managed Chart for GiantSwarm",
			wantTags:         []string{"helmchart", "helmchart-deployable"},
			wantAnnotations: map[string]string{
				"application.giantswarm.io/audience": "giantswarm",
				"application.giantswarm.io/managed":  "true",
				"giantswarm.io/helmcharts":           "gsoci.azurecr.io/giantswarm/managed-chart",
				"giantswarm.io/helmchart-versions":   "v1.5.0",
			},
			wantLinks:             nil,
			wantVersion:           "v1.5.0",
			wantOwner:             "group:custom/unspecified",
			wantGithubProjectSlug: "giantswarm/managed-chart-app",
			wantCreatedTimeSet:    true,
			wantErr:               false,
		},
		{
			name: "Chart with GitHub home URL",
			repo: "giantswarm/hello-world",
			tag:  "v1.0.0",
			configMap: map[string]interface{}{
				"description": "Hello World chart",
				"home":        "https://github.com/giantswarm/hello-world-app",
			},
			namespace:        "default",
			componentType:    "service",
			registryHostname: "gsoci.azurecr.io",
			wantName:         "hello-world-app",
			wantDescription:  "Hello World chart",
			wantTags:         []string{"helmchart", "helmchart-deployable"},
			wantAnnotations: map[string]string{
				"application.giantswarm.io/audience": "all",
				"application.giantswarm.io/managed":  "false",
				"giantswarm.io/helmcharts":           "gsoci.azurecr.io/giantswarm/hello-world",
				"giantswarm.io/helmchart-versions":   "v1.0.0",
			},
			wantLinks:             nil,
			wantVersion:           "v1.0.0",
			wantOwner:             "group:default/unspecified",
			wantGithubProjectSlug: "giantswarm/hello-world-app",
			wantCreatedTimeSet:    true,
			wantErr:               false,
		},
		{
			name: "Chart without home URL should fail",
			repo: "external/some-chart",
			tag:  "v2.0.0",
			configMap: map[string]interface{}{
				"description": "External chart",
			},
			namespace:        "default",
			componentType:    "service",
			registryHostname: "registry.example.com",
			wantErr:          true,
		},
		{
			name: "Chart with non-GitHub home URL should fail",
			repo: "internal/chart",
			tag:  "v1.0.0",
			configMap: map[string]interface{}{
				"description": "Internal chart",
				"home":        "https://internal.example.com/charts/my-chart",
			},
			namespace:        "default",
			componentType:    "service",
			registryHostname: "registry.example.com",
			wantErr:          true,
		},
		{
			name: "Chart with non-giantswarm GitHub home URL should fail",
			repo: "external/chart",
			tag:  "v1.0.0",
			configMap: map[string]interface{}{
				"description": "External chart",
				"home":        "https://github.com/external-org/some-chart",
			},
			namespace:        "default",
			componentType:    "service",
			registryHostname: "registry.example.com",
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createComponentFromOCIChart(
				tt.repo,
				tt.tag,
				tt.configMap,
				tt.namespace,
				tt.componentType,
				tt.registryHostname,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("createComponentFromOCIChart() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return // Skip further checks if error was expected
			}

			// Check basic fields
			if got.Name != tt.wantName {
				t.Errorf("createComponentFromOCIChart() Name = %v, want %v", got.Name, tt.wantName)
			}

			if got.Description != tt.wantDescription {
				t.Errorf("createComponentFromOCIChart() Description = %v, want %v", got.Description, tt.wantDescription)
			}

			if got.Namespace != tt.namespace {
				t.Errorf("createComponentFromOCIChart() Namespace = %v, want %v", got.Namespace, tt.namespace)
			}

			if got.Owner != tt.wantOwner {
				t.Errorf("createComponentFromOCIChart() Owner = %v, want %v", got.Owner, tt.wantOwner)
			}

			if got.Type != tt.componentType {
				t.Errorf("createComponentFromOCIChart() Type = %v, want %v", got.Type, tt.componentType)
			}

			// Check tags
			if diff := cmp.Diff(tt.wantTags, got.Tags); diff != "" {
				t.Errorf("createComponentFromOCIChart() Tags mismatch (-want +got):\n%s", diff)
			}

			// Check annotations
			if diff := cmp.Diff(tt.wantAnnotations, got.Annotations); diff != "" {
				t.Errorf("createComponentFromOCIChart() Annotations mismatch (-want +got):\n%s", diff)
			}

			// Check links
			if diff := cmp.Diff(tt.wantLinks, got.Links); diff != "" {
				t.Errorf("createComponentFromOCIChart() Links mismatch (-want +got):\n%s", diff)
			}

			// Check version
			if got.LatestReleaseTag != tt.wantVersion {
				t.Errorf("createComponentFromOCIChart() LatestReleaseTag = %v, want %v", got.LatestReleaseTag, tt.wantVersion)
			}

			// Check GitHub project slug
			if got.GithubProjectSlug != tt.wantGithubProjectSlug {
				t.Errorf("createComponentFromOCIChart() GithubProjectSlug = %v, want %v", got.GithubProjectSlug, tt.wantGithubProjectSlug)
			}

			// Check that creation time is set
			if tt.wantCreatedTimeSet && got.LatestReleaseTime.IsZero() {
				t.Errorf("createComponentFromOCIChart() LatestReleaseTime should not be zero")
			}

			// For test cases with specific created time, verify it's parsed correctly
			if tt.configMap != nil {
				if created, ok := tt.configMap["created"].(string); ok && created == "2023-10-15T10:30:00Z" {
					if !got.LatestReleaseTime.Equal(fixedTime) {
						t.Errorf("createComponentFromOCIChart() LatestReleaseTime = %v, want %v", got.LatestReleaseTime, fixedTime)
					}
				}
			}
		})
	}
}

// TestCreateComponentFromOCIChart_ErrorCases tests error scenarios
func TestCreateComponentFromOCIChart_ErrorCases(t *testing.T) {
	tests := []struct {
		name             string
		repo             string
		tag              string
		configMap        map[string]interface{}
		namespace        string
		componentType    string
		registryHostname string
		wantErr          bool
		wantErrContains  string
	}{
		{
			name:             "Missing home field should error",
			repo:             "giantswarm/test-chart",
			tag:              "v1.0.0",
			configMap:        nil,
			namespace:        "default",
			componentType:    "service",
			registryHostname: "registry.example.com",
			wantErr:          true,
			wantErrContains:  "cannot match chart to GitHub repository",
		},
		{
			name: "Empty home field should error",
			repo: "giantswarm/test-chart",
			tag:  "v1.0.0",
			configMap: map[string]interface{}{
				"home": "",
			},
			namespace:        "default",
			componentType:    "service",
			registryHostname: "registry.example.com",
			wantErr:          true,
			wantErrContains:  "cannot match chart to GitHub repository",
		},
		{
			name: "Non-GitHub home URL should error",
			repo: "giantswarm/test-chart",
			tag:  "v1.0.0",
			configMap: map[string]interface{}{
				"home": "https://example.com/charts/my-chart",
			},
			namespace:        "default",
			componentType:    "service",
			registryHostname: "registry.example.com",
			wantErr:          true,
			wantErrContains:  "cannot match chart to GitHub repository",
		},
		{
			name: "Non-giantswarm GitHub URL should error",
			repo: "external/test-chart",
			tag:  "v1.0.0",
			configMap: map[string]interface{}{
				"home": "https://github.com/external-org/test-chart",
			},
			namespace:        "default",
			componentType:    "service",
			registryHostname: "registry.example.com",
			wantErr:          true,
			wantErrContains:  "cannot match chart to GitHub repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := createComponentFromOCIChart(
				tt.repo,
				tt.tag,
				tt.configMap,
				tt.namespace,
				tt.componentType,
				tt.registryHostname,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("createComponentFromOCIChart() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.wantErrContains != "" {
				if err == nil || !contains(err.Error(), tt.wantErrContains) {
					t.Errorf("createComponentFromOCIChart() error = %v, want error containing %v", err, tt.wantErrContains)
				}
			}
		})
	}
}

// TestFormatTeamOwner tests the team owner formatting function
func TestFormatTeamOwner(t *testing.T) {
	tests := []struct {
		name      string
		team      string
		namespace string
		expected  string
	}{
		{
			name:      "Team without prefix in giantswarm namespace",
			team:      "honeybadger",
			namespace: "giantswarm",
			expected:  "group:giantswarm/team-honeybadger",
		},
		{
			name:      "Team with prefix in default namespace",
			team:      "team-atlas",
			namespace: "default",
			expected:  "group:default/team-atlas",
		},
		{
			name:      "Team with prefix in production namespace",
			team:      "team-bigmac",
			namespace: "production",
			expected:  "group:production/team-bigmac",
		},
		{
			name:      "Single letter team in custom namespace",
			team:      "a",
			namespace: "custom",
			expected:  "group:custom/team-a",
		},
		{
			name:      "Team with numbers in development namespace",
			team:      "team-123",
			namespace: "development",
			expected:  "group:development/team-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTeamOwner(tt.team, tt.namespace)
			if got != tt.expected {
				t.Errorf("formatTeamOwner(%q, %q) = %q, want %q", tt.team, tt.namespace, got, tt.expected)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
