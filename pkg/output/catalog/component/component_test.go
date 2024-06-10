package component

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/output/bscatalog/v1alpha1"
)

// TestComponent_ToEntity tests the ToEntity method of the Component struct.
// This gives us the opportunity to use all setters and options.
func TestComponent_ToEntity(t *testing.T) {

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
				WithOpsGenieComponentSelector("ops-genie-selector"),
				WithOpsGenieTeam("my-ops-team"),
				WithOwner("my-owner"),
				WithQuayRepositorySlug("namespace/quay-project"),
				WithSystem("my-system"),
				WithTags("tag1", "tag2"),
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
					Labels:      map[string]string{"key": "value"},
					Annotations: map[string]string{
						"backstage.io/kubernetes-id":        "my-k8s-id",
						"backstage.io/source-location":      "url:https://github.com/foo-org/my-project",
						"backstage.io/techdocs-ref":         "url:https://github.com/foo-org/my-project/tree/master",
						"circleci.com/project-slug":         "my-circleci-slug",
						"giantswarm.io/deployment-names":    "name1,name2",
						"giantswarm.io/latest-release-date": "2018-01-03T01:02:03Z",
						"giantswarm.io/latest-release-tag":  "v5.0.1",
						"github.com/project-slug":           "foo-org/my-project",
						"github.com/team-slug":              "my-team",
						"opsgenie.com/component-selector":   "ops-genie-selector",
						"opsgenie.com/team":                 "my-ops-team",
						"quay.io/repository-slug":           "namespace/quay-project",
					},
					Links: []bscatalog.EntityLink{},
					Tags:  []string{"tag1", "tag2"},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component, err := New(tt.componentName, tt.options...)
			if err != nil && !tt.wantErr {
				t.Errorf("NewComponent() error = %v, wantErr %v", err, tt.wantErr)
				return
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
				Name:      "minimal",
				Namespace: "default",
				Owner:     "unspecified",
				Type:      "unspecified",
				Lifecycle: "production",
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
				Name:      "minimal",
				Namespace: "default",
				Owner:     "unspecified",
				Type:      "unspecified",
				Lifecycle: "production",
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
				Name:      "minimal",
				Namespace: "default",
				Owner:     "unspecified",
				Tags:      []string{"tag1"},
				Type:      "unspecified",
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
