package catalog

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestComponent_ToEntity(t *testing.T) {

	tests := []struct {
		name          string
		componentName string
		options       []Option
		want          *Entity
		wantErr       bool
	}{
		{
			name:          "Minimal",
			componentName: "minimal",
			options:       []Option{},
			want: &Entity{
				APIVersion: "backstage.io/v1alpha1",
				Kind:       EntityKindComponent,
				Metadata: EntityMetadata{
					Name:        "minimal",
					Labels:      map[string]string{},
					Annotations: map[string]string{},
					Links:       []EntityLink{},
				},
				Spec: ComponentSpec{
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
				WithDescription("A full-fledged component"),
				WithTitle("Full Fledged"),
				WithGithubTeamSlug("my-team"),
				WithHasReadme(true),
				WithKubernetesID("my-k8s-id"),
				WithLatestReleaseTag("v5.0.1"),
				WithLifecycle("deprecated"),
				WithNamespace("my-namespace"),
				WithOwner("my-owner"),
				WithSystem("my-system"),
				WithTags("tag1", "tag2"),
				WithType("component-type"),
				WithLatestReleaseTime(time.Date(2018, time.January, 3, 1, 2, 3, 0, time.UTC)),
				WithOpsGenieTeam("my-ops-team"),
				WithGithubProjectSlug("foo-org/my-project"),
				WithQuayRepositorySlug("namespace/quay-project"),
				WithDeploymentNames("name1", "name2"),
				WithOpsGenieComponentSelector("ops-genie-selector"),
				WithLabels(map[string]string{"key": "value"}),
			},
			want: &Entity{
				APIVersion: "backstage.io/v1alpha1",
				Kind:       EntityKindComponent,
				Metadata: EntityMetadata{
					Name:        "full-fledged",
					Namespace:   "my-namespace",
					Description: "A full-fledged component",
					Labels:      map[string]string{"key": "value"},
					Annotations: map[string]string{
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
					Links: []EntityLink{},
					Tags:  []string{"tag1", "tag2"},
					Title: "Full Fledged",
				},
				Spec: ComponentSpec{
					Type:      "component-type",
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
			component, err := NewComponent(tt.componentName, tt.options...)
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
