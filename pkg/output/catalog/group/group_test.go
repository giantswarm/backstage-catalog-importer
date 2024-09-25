package group

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/output/bscatalog/v1alpha1"
)

func TestGroup_ToEntity(t *testing.T) {
	tests := []struct {
		name      string
		groupName string
		options   []Option
		want      *bscatalog.Entity
		wantErr   bool
	}{
		{
			name:      "Minimal",
			groupName: "minimal-group",
			options:   []Option{},
			want: &bscatalog.Entity{
				APIVersion: "backstage.io/v1alpha1",
				Kind:       bscatalog.EntityKindGroup,
				Metadata: bscatalog.EntityMetadata{
					Name:      "minimal-group",
					Namespace: "default",
				},
				Spec: bscatalog.GroupSpec{
					Type: "team",
				},
			},
			wantErr: false,
		},
		{
			name:      "Fullfledged",
			groupName: "full-fledged-group",
			options: []Option{
				WithNamespace("namespace"),
				WithTitle("Full Fledged"),
				WithDescription("A full-fledged group"),
				WithPictureURL("https://example.com/picture.jpg"),
				WithGrafanaDashboardSelector("my-dashboard"),
				WithChildrenNames("child2", "child1"),
				WithParentName("parent"),
				WithMemberNames("member2", "member1"),
			},
			want: &bscatalog.Entity{
				APIVersion: "backstage.io/v1alpha1",
				Kind:       bscatalog.EntityKindGroup,
				Metadata: bscatalog.EntityMetadata{
					Name:        "full-fledged-group",
					Description: "A full-fledged group",
					Title:       "Full Fledged",
					Namespace:   "namespace",
					Annotations: map[string]string{
						"grafana/dashboard-selector": "my-dashboard",
					},
				},
				Spec: bscatalog.GroupSpec{
					Type: "team",
					Profile: bscatalog.GroupProfile{
						DisplayName: "Full Fledged",
						Picture:     "https://example.com/picture.jpg",
					},
					Children: []string{"child1", "child2"},
					Parent:   "parent",
					Members:  []string{"member1", "member2"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := New(tt.groupName, tt.options...)
			if err != nil && !tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got := c.ToEntity()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Group.ToEntity() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
