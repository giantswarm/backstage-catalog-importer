package export

import (
	"flag"
	"io"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/bscatalog/v1alpha1"
)

var (
	update = flag.Bool("update", false, "update the golden files of this test")
)

// TestServiceOutput is a quite deep test that also covers the entity generation.
// Make sure that it covers all entity kinds and fields.
func TestServiceOutput(t *testing.T) {
	tests := []struct {
		name       string
		goldenFile string
		entities   []bscatalog.Entity
	}{
		{
			name:       "Group, users, and component",
			goldenFile: "case01.golden",
			entities: []bscatalog.Entity{
				{
					APIVersion: bscatalog.APIVersion,
					Kind:       bscatalog.EntityKindGroup,
					Metadata: bscatalog.EntityMetadata{
						Name:        "myorg/team-slug",
						Description: "A simple team with simple people",
						Annotations: map[string]string{
							"grafana/dashboard-selector": "tags @> 'owner:myorg/team-slug'",
							"opsgenie.com/team":          "myorg/team-slug",
						},
					},
					Spec: bscatalog.GroupSpec{
						Type: "team",
						Profile: bscatalog.GroupProfile{
							DisplayName: "team-name",
							Picture:     "https://avatars.githubusercontent.com/t/16638849?s=116&v=4",
						},
						Parent:  "area-everything",
						Members: []string{"jane-doe", "second-member"},
					},
				},
				{
					APIVersion: bscatalog.APIVersion,
					Kind:       bscatalog.EntityKindUser,
					Metadata: bscatalog.EntityMetadata{
						Name:        "jane-doe",
						Description: "Experienced DevOps engineer, jack of all trades",
					},
					Spec: bscatalog.UserSpec{
						Profile: bscatalog.UserProfile{
							Email:       "jane@acme.org",
							DisplayName: "Jane Doe",
							Picture:     "https://avatars.githubusercontent.com/u/12345678?v=4",
						},
					},
				},
				{
					APIVersion: bscatalog.APIVersion,
					Kind:       bscatalog.EntityKindComponent,
					Metadata: bscatalog.EntityMetadata{
						Name:        "my-service",
						Description: "Awesome microservice",
						Annotations: map[string]string{
							"backstage.io/kubernetes-id":           "my-service",
							"backstage.io/source-location":         "url:https://github.com/giantswarm/my-service",
							"backstage.io/techdocs-ref":            "url:https://github.com/giantswarm/my-service/tree/main",
							"circleci.com/project-slug":            "github/giantswarm/my-service",
							"giantswarm.io/deployment-names":       "my-service,my-service-app",
							"giantswarm.io/helmchart-app-versions": ",2.3.4",
							"giantswarm.io/helmchart-versions":     "1.2.3,0.4.1",
							"giantswarm.io/helmcharts":             "first-chart,second-chart",
							"giantswarm.io/latest-release-tag":     "v1.2.3",
							"github.com/project-slug":              "giantswarm/my-service",
							"github.com/team-slug":                 "myorg/team-slug",
							"opsgenie.com/component-selector":      "detailsPair(app:my-service) OR detailsPair(app:my-service-app)",
							"opsgenie.com/team":                    "myorg/team-slug",
							"quay.io/repository-slug":              "giantswarm/my-service",
						},
						Labels: map[string]string{
							"giantswarm.io/flavor-app": "true",
							"giantswarm.io/language":   "go",
						},
						Tags: []string{"flavor:app", "helmchart", "language:go"},
						Links: []bscatalog.EntityLink{
							{
								URL:   "https://giantswarm.grafana.net/d/eb617ba1-209a-4d57-9963-1af9a8ddc8d4/general-service-metrics?orgId=1&var-app=my-service&var-app=my-service-app&from=now-24h&to=now",
								Title: "General service metrics dashboard",
								Icon:  "dashboard",
								Type:  "grafana-dashboard",
							},
						},
					},
					Spec: bscatalog.ComponentSpec{
						Type:      "service",
						Lifecycle: "production",
						Owner:     "myorg/team-slug",
						System:    "everything-system",
						DependsOn: []string{"component:first-dependency", "component:second-dependency"},
					},
				},
			},
		},
		{
			name:       "Component with individual deployment names",
			goldenFile: "case02.golden",
			entities: []bscatalog.Entity{
				{
					APIVersion: bscatalog.APIVersion,
					Kind:       bscatalog.EntityKindComponent,
					Metadata: bscatalog.EntityMetadata{
						Name:        "project-with-two-apps",
						Description: "Project that includes two apps",
						Annotations: map[string]string{
							"backstage.io/kubernetes-id":      "project-with-two-apps",
							"backstage.io/source-location":    "url:https://github.com/giantswarm/project-with-two-apps",
							"backstage.io/techdocs-ref":       "url:https://github.com/giantswarm/project-with-two-apps/tree/master",
							"circleci.com/project-slug":       "github/giantswarm/project-with-two-apps",
							"giantswarm.io/deployment-names":  "first-name,second-name-app",
							"github.com/project-slug":         "giantswarm/project-with-two-apps",
							"github.com/team-slug":            "myorg/team-slug",
							"opsgenie.com/component-selector": "detailsPair(app:first-name) OR detailsPair(app:second-name-app)",
							"opsgenie.com/team":               "myorg/team-slug",
							"quay.io/repository-slug":         "giantswarm/project-with-two-apps",
						},
					},
					Spec: bscatalog.ComponentSpec{
						Type:      "service",
						Lifecycle: "production",
						Owner:     "myorg/team-slug",
						System:    "everything-system",
						DependsOn: []string{"component:first-dependency", "component:second-dependency"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(Config{
				TargetPath: ".",
			})

			for idx := range tt.entities {
				err := s.AddEntity(&tt.entities[idx])
				if err != nil {
					t.Errorf("TestServiceOutput - Unexpected error in AddEntity(): %v", err)
				}
			}

			got := s.String()
			want := goldenValue(t, tt.goldenFile, got, *update)
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Service.Bytes() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func goldenValue(t *testing.T, goldenFile string, actual string, update bool) string {
	t.Helper()
	goldenPath := "testdata/" + goldenFile

	f, err := os.OpenFile(goldenPath, os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("Error opening file %s: %s", goldenPath, err)
	}
	defer f.Close()

	if update {
		_, err := f.WriteString(actual)
		if err != nil {
			t.Fatalf("Error writing to file %s: %s", goldenPath, err)
		}

		return actual
	}

	content, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("Error opening file %s: %s", goldenPath, err)
	}
	return string(content)
}
