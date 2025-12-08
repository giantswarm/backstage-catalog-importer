package component

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/backstage-catalog-importer/pkg/input/helmchart"
	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/output/bscatalog/v1alpha1"
)

// TestComponent_ToEntity tests the ToEntity method of the Component struct.
// This gives us the opportunity to use all setters and options.
func TestComponent_ToEntity(t *testing.T) {
	// Create mock helm charts for testing
	mockChart1 := &helmchart.Chart{}
	mockChart1.Name = "first-chart"
	mockChart1.Version = "1.2.3"
	mockChart1.AppVersion = ""

	mockChart2 := &helmchart.Chart{}
	mockChart2.Name = "second-chart"
	mockChart2.Version = "0.4.1"
	mockChart2.AppVersion = "2.3.4"

	tests := []struct {
		name          string
		componentName string
		options       []Option
		want          *bscatalog.Entity
		wantErr       bool
	}{
		{
			name:          "Minimal",
			componentName: "minimal",
			options:       []Option{},
			want: &bscatalog.Entity{
				APIVersion: bscatalog.APIVersion,
				Kind:       bscatalog.EntityKindComponent,
				Metadata: bscatalog.EntityMetadata{
					Name:        "minimal",
					Labels:      map[string]string{},
					Annotations: map[string]string{},
					Links:       []bscatalog.EntityLink{},
				},
				Spec: bscatalog.ComponentSpec{
					Type:      "unspecified",
					Lifecycle: "production",
					Owner:     "unspecified",
				},
			},
			wantErr: false,
		},
		{
			name:          "Fullfledged",
			componentName: "full-fledged",
			options: []Option{
				WithLanguage("go"),
				WithPrivate(true),
				WithFlavors("app"),
				WithHelmCharts(mockChart1, mockChart2),
				WithCircleCiSlug("my-circleci-slug"),
				WithDefaultBranch("master"),
				WithDependsOn("first-dependency", "second-dependency"),
				WithDeploymentNames("name1", "name2"),
				WithDescription("A full-fledged component"),
				WithGithubProjectSlug("foo-org/my-project"),
				WithGithubTeamSlug("my-team"),
				WithHasReadme(true),
				WithKubernetesID("my-k8s-id"),
				WithLabels(map[string]string{"key": "value"}),
				WithLatestReleaseTag("v5.0.1"),
				WithLatestReleaseTime(time.Date(2018, time.January, 3, 1, 2, 3, 0, time.UTC)),
				WithLifecycle("deprecated"),
				WithNamespace("my-namespace"),
				WithOwner("my-owner"),
				WithSystem("my-system"),
				WithTags("My furious tag 1", "SuperBad_2"),
				WithTitle("Full Fledged"),
				WithType("service"),
			},
			want: &bscatalog.Entity{
				APIVersion: bscatalog.APIVersion,
				Kind:       bscatalog.EntityKindComponent,
				Metadata: bscatalog.EntityMetadata{
					Name:        "full-fledged",
					Namespace:   "my-namespace",
					Description: "A full-fledged component",
					Labels: map[string]string{
						"key":                      "value",
						"giantswarm.io/language":   "go",
						"giantswarm.io/flavor-app": "true",
					},
					Annotations: map[string]string{
						"backstage.io/kubernetes-id":           "my-k8s-id",
						"backstage.io/source-location":         "url:https://github.com/foo-org/my-project",
						"backstage.io/techdocs-ref":            "url:https://github.com/foo-org/my-project/tree/master",
						"circleci.com/project-slug":            "my-circleci-slug",
						"giantswarm.io/deployment-names":       "name1,name2",
						"giantswarm.io/latest-release-date":    "2018-01-03T01:02:03Z",
						"giantswarm.io/latest-release-tag":     "v5.0.1",
						"github.com/project-slug":              "foo-org/my-project",
						"github.com/team-slug":                 "my-team",
						"giantswarm.io/helmcharts":             "first-chart,second-chart",
						"giantswarm.io/helmchart-versions":     "1.2.3,0.4.1",
						"giantswarm.io/helmchart-app-versions": ",2.3.4",
					},
					Links: []bscatalog.EntityLink{},
					Tags:  []string{"my-furious-tag-1", "superbad-2", "defaultbranch:master", "flavor:app", "helmchart", "private", "language:go"},
					Title: "Full Fledged",
				},
				Spec: bscatalog.ComponentSpec{
					Type:      "service",
					Lifecycle: "deprecated",
					Owner:     "my-owner",
					System:    "my-system",
					DependsOn: []string{"component:first-dependency", "component:second-dependency"},
				},
			},
			wantErr: false,
		},
		{
			name:          "NoReleases",
			componentName: "no-releases-component",
			options: []Option{
				WithHasReleases(false),
				WithLanguage("python"),
				WithFlavors("cli"),
			},
			want: &bscatalog.Entity{
				APIVersion: bscatalog.APIVersion,
				Kind:       bscatalog.EntityKindComponent,
				Metadata: bscatalog.EntityMetadata{
					Name: "no-releases-component",
					Labels: map[string]string{
						"giantswarm.io/language":   "python",
						"giantswarm.io/flavor-cli": "true",
					},
					Annotations: map[string]string{},
					Links:       []bscatalog.EntityLink{},
					Tags:        []string{"flavor:cli", "no-releases", "language:python"},
				},
				Spec: bscatalog.ComponentSpec{
					Type:      "unspecified",
					Lifecycle: "production",
					Owner:     "unspecified",
				},
			},
			wantErr: false,
		},
		{
			name:          "ServiceWithoutKubernetesID",
			componentName: "my-service",
			options: []Option{
				WithType("service"),
			},
			want: &bscatalog.Entity{
				APIVersion: bscatalog.APIVersion,
				Kind:       bscatalog.EntityKindComponent,
				Metadata: bscatalog.EntityMetadata{
					Name:   "my-service",
					Labels: map[string]string{},
					Annotations: map[string]string{
						"backstage.io/kubernetes-id": "my-service",
					},
					Links: []bscatalog.EntityLink{},
				},
				Spec: bscatalog.ComponentSpec{
					Type:      "service",
					Lifecycle: "production",
					Owner:     "unspecified",
				},
			},
			wantErr: false,
		},
		{
			name:          "WithCustomAnnotationsAndLinks",
			componentName: "custom-component",
			options:       []Option{},
			want: &bscatalog.Entity{
				APIVersion: bscatalog.APIVersion,
				Kind:       bscatalog.EntityKindComponent,
				Metadata: bscatalog.EntityMetadata{
					Name:   "custom-component",
					Labels: map[string]string{},
					Annotations: map[string]string{
						"custom.io/annotation": "custom-value",
					},
					Links: []bscatalog.EntityLink{
						{
							URL:   "https://custom.example.com",
							Title: "Custom Link",
							Icon:  "link",
							Type:  "custom",
						},
					},
				},
				Spec: bscatalog.ComponentSpec{
					Type:      "unspecified",
					Lifecycle: "production",
					Owner:     "unspecified",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component, err := New(tt.componentName, tt.options...)
			if err != nil && !tt.wantErr {
				t.Errorf("NewComponent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Apply custom modifications for specific test cases
			if tt.name == "WithCustomAnnotationsAndLinks" {
				component.SetAnnotation("custom.io/annotation", "custom-value")
				component.AddLink(bscatalog.EntityLink{
					URL:   "https://custom.example.com",
					Title: "Custom Link",
					Icon:  "link",
					Type:  "custom",
				})
			}

			got := component.ToEntity()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Component.ToEntity() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		name    string
		options []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *Component
		wantErr bool
	}{
		{
			name: "Success",
			args: args{name: "minimal"},
			want: &Component{
				Name:        "minimal",
				Namespace:   "default",
				Owner:       "unspecified",
				Type:        "unspecified",
				Lifecycle:   "production",
				HasReleases: true,
			},
		},
		{
			name:    "Empty name",
			args:    args{name: ""},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.name, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Component.New() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// Tests Component creation, adders, setters, and options in more flexible ways.
func TestGeneric(t *testing.T) {
	tests := []struct {
		name    string
		code    func() (*Component, error)
		want    *Component
		wantErr bool
	}{
		{
			name: "Simple success",
			code: func() (*Component, error) {
				return New("minimal")
			},
			want: &Component{
				Name:        "minimal",
				Namespace:   "default",
				Owner:       "unspecified",
				Type:        "unspecified",
				Lifecycle:   "production",
				HasReleases: true,
			},
		},
		{
			name: "Some setters",
			code: func() (*Component, error) {
				c, _ := New("minimal")
				c.AddTag("tag1")
				c.AddLink(bscatalog.EntityLink{
					Title: "link1",
					Type:  "dashboard",
					URL:   "https://example.com",
					Icon:  "dashboard",
				})
				c.SetAnnotation("key1", "value1")
				c.SetLabel("label1", "value1")
				return c, nil
			},
			want: &Component{
				Annotations: map[string]string{"key1": "value1"},
				Labels:      map[string]string{"label1": "value1"},
				Lifecycle:   "production",
				Links: []bscatalog.EntityLink{
					{URL: "https://example.com", Title: "link1", Icon: "dashboard", Type: "dashboard"},
				},
				Name:        "minimal",
				Namespace:   "default",
				Owner:       "unspecified",
				Tags:        []string{"tag1"},
				Type:        "unspecified",
				HasReleases: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.code()
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Component.New() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
