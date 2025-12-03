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
			name:             "Simple OCI chart with minimal data",
			repo:             "giantswarm/my-chart",
			tag:              "v1.0.0",
			configMap:        nil,
			namespace:        "default",
			componentType:    "service",
			registryHostname: "registry.example.com",
			wantName:         "my-chart",
			wantDescription:  "OCI chart from giantswarm/my-chart",
			wantTags:         []string{"oci", "helm-chart"},
			wantAnnotations: map[string]string{
				"giantswarm.io/oci-registry":   "registry.example.com",
				"giantswarm.io/oci-repository": "giantswarm/my-chart",
				"giantswarm.io/oci-tag":        "v1.0.0",
			},
			wantLinks:          nil,
			wantVersion:        "v1.0.0",
			wantOwner:          "group:default/unspecified",
			wantCreatedTimeSet: true,
			wantErr:            false,
		},
		{
			name: "OCI chart with full metadata",
			repo: "giantswarm/advanced-chart",
			tag:  "v2.1.0",
			configMap: map[string]interface{}{
				"description": "Advanced Helm chart for testing",
				"version":     "2.1.0",
				"created":     "2023-10-15T10:30:00Z",
			},
			namespace:        "giantswarm",
			componentType:    "library",
			registryHostname: "localhost:5000",
			wantName:         "advanced-chart",
			wantDescription:  "Advanced Helm chart for testing",
			wantTags:         []string{"oci", "helm-chart"},
			wantAnnotations: map[string]string{
				"giantswarm.io/oci-registry":   "localhost:5000",
				"giantswarm.io/oci-repository": "giantswarm/advanced-chart",
				"giantswarm.io/oci-tag":        "v2.1.0",
			},
			wantLinks:          nil,
			wantVersion:        "2.1.0", // Should use version from config
			wantOwner:          "group:giantswarm/unspecified",
			wantCreatedTimeSet: true,
			wantErr:            false,
		},
		{
			name:             "Single name repository",
			repo:             "simple-chart",
			tag:              "latest",
			configMap:        nil,
			namespace:        "default",
			componentType:    "service",
			registryHostname: "docker.io",
			wantName:         "simple-chart",
			wantDescription:  "OCI chart from simple-chart",
			wantTags:         []string{"oci", "helm-chart"},
			wantAnnotations: map[string]string{
				"giantswarm.io/oci-registry":   "docker.io",
				"giantswarm.io/oci-repository": "simple-chart",
				"giantswarm.io/oci-tag":        "latest",
			},
			wantLinks:          nil,
			wantVersion:        "latest",
			wantOwner:          "group:default/unspecified",
			wantCreatedTimeSet: true,
			wantErr:            false,
		},
		{
			name: "Chart with description from config",
			repo: "org/chart-with-description",
			tag:  "v1.5.0",
			configMap: map[string]interface{}{
				"description": "My Awesome Chart",
			},
			namespace:        "custom",
			componentType:    "resource",
			registryHostname: "gsoci.azurecr.io",
			wantName:         "chart-with-description",
			wantDescription:  "My Awesome Chart",
			wantTags:         []string{"oci", "helm-chart"},
			wantAnnotations: map[string]string{
				"giantswarm.io/oci-registry":   "gsoci.azurecr.io",
				"giantswarm.io/oci-repository": "org/chart-with-description",
				"giantswarm.io/oci-tag":        "v1.5.0",
			},
			wantLinks:          nil,
			wantVersion:        "v1.5.0",
			wantOwner:          "group:custom/unspecified",
			wantCreatedTimeSet: true,
			wantErr:            false,
		},
		{
			name: "Chart with empty metadata values",
			repo: "test/empty-metadata",
			tag:  "v0.1.0",
			configMap: map[string]interface{}{
				"description": "", // Empty description
				"version":     "", // Empty version
			},
			namespace:        "default",
			componentType:    "service",
			registryHostname: "registry.test.com",
			wantName:         "empty-metadata",
			wantDescription:  "OCI chart from test/empty-metadata", // Should fall back to default
			wantTags:         []string{"oci", "helm-chart"},
			wantAnnotations: map[string]string{
				"giantswarm.io/oci-registry":   "registry.test.com",
				"giantswarm.io/oci-repository": "test/empty-metadata",
				"giantswarm.io/oci-tag":        "v0.1.0",
			},
			wantLinks:          nil,
			wantVersion:        "v0.1.0", // Should use tag since version is empty
			wantOwner:          "group:default/unspecified",
			wantCreatedTimeSet: true,
			wantErr:            false,
		},
		{
			name: "Chart with localhost registry (HTTP)",
			repo: "local/test-chart",
			tag:  "dev",
			configMap: map[string]interface{}{
				"created": "2023-12-01T15:45:30Z",
			},
			namespace:        "development",
			componentType:    "service",
			registryHostname: "localhost:5000",
			wantName:         "test-chart",
			wantDescription:  "OCI chart from local/test-chart",
			wantTags:         []string{"oci", "helm-chart"},
			wantAnnotations: map[string]string{
				"giantswarm.io/oci-registry":   "localhost:5000",
				"giantswarm.io/oci-repository": "local/test-chart",
				"giantswarm.io/oci-tag":        "dev",
			},
			wantLinks:          nil,
			wantVersion:        "dev",
			wantOwner:          "group:development/unspecified",
			wantCreatedTimeSet: true,
			wantErr:            false,
		},
		{
			name: "Chart with registry without dots (HTTP)",
			repo: "internal/chart",
			tag:  "v2.0.0",
			configMap: map[string]interface{}{
				"description": "Internal chart",
			},
			namespace:        "internal",
			componentType:    "service",
			registryHostname: "registry", // No dots, should use HTTP
			wantName:         "chart",
			wantDescription:  "Internal chart",
			wantTags:         []string{"oci", "helm-chart"},
			wantAnnotations: map[string]string{
				"giantswarm.io/oci-registry":   "registry",
				"giantswarm.io/oci-repository": "internal/chart",
				"giantswarm.io/oci-tag":        "v2.0.0",
			},
			wantLinks:          nil,
			wantVersion:        "v2.0.0",
			wantOwner:          "group:internal/unspecified",
			wantCreatedTimeSet: true,
			wantErr:            false,
		},
		{
			name: "Chart with complex nested repository path",
			repo: "org/team/subteam/complex-chart",
			tag:  "v3.2.1",
			configMap: map[string]interface{}{
				"description": "Complex nested chart",
				"version":     "3.2.1-beta",
				"created":     "2023-11-20T09:15:45Z",
			},
			namespace:        "production",
			componentType:    "library",
			registryHostname: "prod.registry.example.com",
			wantName:         "complex-chart", // Should extract last part
			wantDescription:  "Complex nested chart",
			wantTags:         []string{"oci", "helm-chart"},
			wantAnnotations: map[string]string{
				"giantswarm.io/oci-registry":   "prod.registry.example.com",
				"giantswarm.io/oci-repository": "org/team/subteam/complex-chart",
				"giantswarm.io/oci-tag":        "v3.2.1",
			},
			wantLinks:          nil,
			wantVersion:        "3.2.1-beta", // Should use version from config
			wantOwner:          "group:production/unspecified",
			wantCreatedTimeSet: true,
			wantErr:            false,
		},
		{
			name: "Chart with invalid created time format",
			repo: "test/invalid-time",
			tag:  "v1.0.0",
			configMap: map[string]interface{}{
				"created": "invalid-time-format", // Invalid time format
			},
			namespace:        "default",
			componentType:    "service",
			registryHostname: "registry.example.com",
			wantName:         "invalid-time",
			wantDescription:  "OCI chart from test/invalid-time",
			wantTags:         []string{"oci", "helm-chart"},
			wantAnnotations: map[string]string{
				"giantswarm.io/oci-registry":   "registry.example.com",
				"giantswarm.io/oci-repository": "test/invalid-time",
				"giantswarm.io/oci-tag":        "v1.0.0",
			},
			wantLinks:          nil,
			wantVersion:        "v1.0.0",
			wantOwner:          "group:default/unspecified",
			wantCreatedTimeSet: true, // Should still set time to now()
			wantErr:            false,
		},
		{
			name: "Chart with malformed config structure",
			repo: "test/malformed-config",
			tag:  "v1.0.0",
			configMap: map[string]interface{}{
				"config": "not-a-map", // Invalid config structure
			},
			namespace:        "default",
			componentType:    "service",
			registryHostname: "registry.example.com",
			wantName:         "malformed-config",
			wantDescription:  "OCI chart from test/malformed-config", // Should fall back to default
			wantTags:         []string{"oci", "helm-chart"},
			wantAnnotations: map[string]string{
				"giantswarm.io/oci-registry":   "registry.example.com",
				"giantswarm.io/oci-repository": "test/malformed-config",
				"giantswarm.io/oci-tag":        "v1.0.0",
			},
			wantLinks:          nil,
			wantVersion:        "v1.0.0",
			wantOwner:          "group:default/unspecified",
			wantCreatedTimeSet: true,
			wantErr:            false,
		},
		{
			name: "Chart with icon from config",
			repo: "giantswarm/chart-with-icon",
			tag:  "v1.0.0",
			configMap: map[string]interface{}{
				"description": "Chart with icon",
				"icon":        "https://example.com/icon.png",
			},
			namespace:        "default",
			componentType:    "service",
			registryHostname: "registry.example.com",
			wantName:         "chart-with-icon",
			wantDescription:  "Chart with icon",
			wantTags:         []string{"oci", "helm-chart"},
			wantAnnotations: map[string]string{
				"giantswarm.io/oci-registry":   "registry.example.com",
				"giantswarm.io/oci-repository": "giantswarm/chart-with-icon",
				"giantswarm.io/oci-tag":        "v1.0.0",
				"giantswarm.io/icon-url":       "https://example.com/icon.png",
			},
			wantLinks:          nil,
			wantVersion:        "v1.0.0",
			wantOwner:          "group:default/unspecified",
			wantCreatedTimeSet: true,
			wantErr:            false,
		},
		{
			name: "Chart with team owner (without team- prefix)",
			repo: "giantswarm/honeybadger-chart",
			tag:  "v1.0.0",
			configMap: map[string]interface{}{
				"description": "Chart owned by honeybadger team",
				"annotations": map[string]interface{}{
					"application.giantswarm.io/team": "honeybadger",
				},
			},
			namespace:        "default",
			componentType:    "service",
			registryHostname: "gsoci.azurecr.io",
			wantName:         "honeybadger-chart",
			wantDescription:  "Chart owned by honeybadger team",
			wantTags:         []string{"oci", "helm-chart"},
			wantAnnotations: map[string]string{
				"giantswarm.io/oci-registry":   "gsoci.azurecr.io",
				"giantswarm.io/oci-repository": "giantswarm/honeybadger-chart",
				"giantswarm.io/oci-tag":        "v1.0.0",
			},
			wantLinks:          nil,
			wantVersion:        "v1.0.0",
			wantOwner:          "group:default/team-honeybadger",
			wantCreatedTimeSet: true,
			wantErr:            false,
		},
		{
			name: "Chart with team owner (with team- prefix)",
			repo: "giantswarm/atlas-chart",
			tag:  "v2.0.0",
			configMap: map[string]interface{}{
				"description": "Chart owned by team-atlas",
				"annotations": map[string]interface{}{
					"application.giantswarm.io/team": "team-atlas",
				},
			},
			namespace:        "default",
			componentType:    "service",
			registryHostname: "registry.example.com",
			wantName:         "atlas-chart",
			wantDescription:  "Chart owned by team-atlas",
			wantTags:         []string{"oci", "helm-chart"},
			wantAnnotations: map[string]string{
				"giantswarm.io/oci-registry":   "registry.example.com",
				"giantswarm.io/oci-repository": "giantswarm/atlas-chart",
				"giantswarm.io/oci-tag":        "v2.0.0",
			},
			wantLinks:          nil,
			wantVersion:        "v2.0.0",
			wantOwner:          "group:default/team-atlas",
			wantCreatedTimeSet: true,
			wantErr:            false,
		},
		{
			name: "Chart with empty team annotation",
			repo: "giantswarm/no-team-chart",
			tag:  "v1.0.0",
			configMap: map[string]interface{}{
				"description": "Chart with empty team",
				"annotations": map[string]interface{}{
					"application.giantswarm.io/team": "",
				},
			},
			namespace:        "default",
			componentType:    "service",
			registryHostname: "registry.example.com",
			wantName:         "no-team-chart",
			wantDescription:  "Chart with empty team",
			wantTags:         []string{"oci", "helm-chart"},
			wantAnnotations: map[string]string{
				"giantswarm.io/oci-registry":   "registry.example.com",
				"giantswarm.io/oci-repository": "giantswarm/no-team-chart",
				"giantswarm.io/oci-tag":        "v1.0.0",
			},
			wantLinks:          nil,
			wantVersion:        "v1.0.0",
			wantOwner:          "group:default/unspecified", // Should use default owner
			wantCreatedTimeSet: true,
			wantErr:            false,
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
			wantName:         "hello-world",
			wantDescription:  "Hello World chart",
			wantTags:         []string{"oci", "helm-chart"},
			wantAnnotations: map[string]string{
				"giantswarm.io/oci-registry":   "gsoci.azurecr.io",
				"giantswarm.io/oci-repository": "giantswarm/hello-world",
				"giantswarm.io/oci-tag":        "v1.0.0",
			},
			wantLinks:             nil,
			wantVersion:           "v1.0.0",
			wantOwner:             "group:default/unspecified",
			wantGithubProjectSlug: "giantswarm/hello-world-app",
			wantCreatedTimeSet:    true,
			wantErr:               false,
		},
		{
			name: "Chart with non-giantswarm GitHub home URL",
			repo: "external/some-chart",
			tag:  "v2.0.0",
			configMap: map[string]interface{}{
				"description": "External chart",
				"home":        "https://github.com/external-org/some-chart",
			},
			namespace:        "default",
			componentType:    "service",
			registryHostname: "registry.example.com",
			wantName:         "some-chart",
			wantDescription:  "External chart",
			wantTags:         []string{"oci", "helm-chart"},
			wantAnnotations: map[string]string{
				"giantswarm.io/oci-registry":   "registry.example.com",
				"giantswarm.io/oci-repository": "external/some-chart",
				"giantswarm.io/oci-tag":        "v2.0.0",
			},
			wantLinks:             nil,
			wantVersion:           "v2.0.0",
			wantOwner:             "group:default/unspecified",
			wantGithubProjectSlug: "", // Should not be set for non-giantswarm URLs
			wantCreatedTimeSet:    true,
			wantErr:               false,
		},
		{
			name: "Chart with non-GitHub home URL",
			repo: "internal/chart",
			tag:  "v1.0.0",
			configMap: map[string]interface{}{
				"description": "Internal chart",
				"home":        "https://internal.example.com/charts/my-chart",
			},
			namespace:        "default",
			componentType:    "service",
			registryHostname: "registry.example.com",
			wantName:         "chart",
			wantDescription:  "Internal chart",
			wantTags:         []string{"oci", "helm-chart"},
			wantAnnotations: map[string]string{
				"giantswarm.io/oci-registry":   "registry.example.com",
				"giantswarm.io/oci-repository": "internal/chart",
				"giantswarm.io/oci-tag":        "v1.0.0",
			},
			wantLinks:             nil,
			wantVersion:           "v1.0.0",
			wantOwner:             "group:default/unspecified",
			wantGithubProjectSlug: "", // Should not be set for non-GitHub URLs
			wantCreatedTimeSet:    true,
			wantErr:               false,
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
			name:             "Empty repository name",
			repo:             "",
			tag:              "v1.0.0",
			configMap:        nil,
			namespace:        "default",
			componentType:    "service",
			registryHostname: "registry.example.com",
			wantErr:          true,
			wantErrContains:  "name must not be empty",
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
