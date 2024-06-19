package user

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/output/bscatalog/v1alpha1"
)

func TestUser_ToEntity(t *testing.T) {
	tests := []struct {
		name     string
		userName string
		options  []Option
		want     *bscatalog.Entity
		wantErr  bool
	}{
		{
			name:     "Minimal",
			userName: "minimal-user",
			options:  []Option{},
			want: &bscatalog.Entity{
				APIVersion: "backstage.io/v1alpha1",
				Kind:       bscatalog.EntityKindUser,
				Metadata: bscatalog.EntityMetadata{
					Name:      "minimal-user",
					Namespace: "default",
				},
				Spec: bscatalog.UserSpec{},
			},
			wantErr: false,
		},
		{
			name:     "Fullfledged",
			userName: "full-fledged-user",
			options: []Option{
				WithNamespace("namespace"),
				WithTitle("Full Fledged"),
				WithEmail("mail@example.com"),
				WithDescription("A full-fledged user"),
				WithPictureURL("https://example.com/picture.jpg"),
				WithGroups("group2", "group1"),
			},
			want: &bscatalog.Entity{
				APIVersion: "backstage.io/v1alpha1",
				Kind:       bscatalog.EntityKindUser,
				Metadata: bscatalog.EntityMetadata{
					Name:        "full-fledged-user",
					Description: "A full-fledged user",
					Title:       "Full Fledged",
					Namespace:   "namespace",
				},
				Spec: bscatalog.UserSpec{
					Profile: bscatalog.UserProfile{
						DisplayName: "Full Fledged",
						Picture:     "https://example.com/picture.jpg",
						Email:       "mail@example.com",
					},
					MemberOf: []string{"group1", "group2"},
				},
			},
			wantErr: false,
		},
		{
			name:     "EmptyName",
			userName: "",
			options:  []Option{},
			want:     nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := New(tt.userName, tt.options...)
			if err != nil && !tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if c != nil {
				got := c.ToEntity()
				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("User.ToEntity() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
