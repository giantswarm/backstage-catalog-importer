package groups

import (
	"reflect"
	"testing"

	"github.com/google/go-github/v88/github"
)

func TestGroupFromTeam(t *testing.T) {
	testCases := []struct {
		name        string
		team        *github.Team
		memberNames []string

		expectedName     string
		expectedTitle    string
		expectedDesc     string
		expectedPicture  string
		expectedParent   string
		expectedSelector string
		expectedMembers  []string
	}{
		{
			name: "team without parent and with members",
			team: &github.Team{
				ID:          github.Ptr(int64(123)),
				Slug:        github.Ptr("team-honeybadger"),
				Name:        github.Ptr("Honey Badger"),
				Description: github.Ptr("The honey badger team"),
			},
			memberNames:      []string{"bob", "alice"},
			expectedName:     "team-honeybadger",
			expectedTitle:    "Honey Badger",
			expectedDesc:     "The honey badger team",
			expectedPicture:  "https://avatars.githubusercontent.com/t/123?s=116&v=4",
			expectedParent:   "",
			expectedSelector: "tags @> 'owner:team-honeybadger'",
			expectedMembers:  []string{"bob", "alice"},
		},
		{
			name: "team with parent and no members",
			team: &github.Team{
				ID:          github.Ptr(int64(456)),
				Slug:        github.Ptr("team-bumblebee"),
				Name:        github.Ptr("Bumblebee"),
				Description: github.Ptr(""),
				Parent: &github.Team{
					Slug: github.Ptr("team-parent"),
				},
			},
			memberNames:      nil,
			expectedName:     "team-bumblebee",
			expectedTitle:    "Bumblebee",
			expectedDesc:     "",
			expectedPicture:  "https://avatars.githubusercontent.com/t/456?s=116&v=4",
			expectedParent:   "team-parent",
			expectedSelector: "tags @> 'owner:team-bumblebee'",
			expectedMembers:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g, err := groupFromTeam(tc.team, tc.memberNames)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if g.Name != tc.expectedName {
				t.Errorf("Name: got %q, want %q", g.Name, tc.expectedName)
			}
			if g.Title != tc.expectedTitle {
				t.Errorf("Title: got %q, want %q", g.Title, tc.expectedTitle)
			}
			if g.Description != tc.expectedDesc {
				t.Errorf("Description: got %q, want %q", g.Description, tc.expectedDesc)
			}
			if g.PictureURL != tc.expectedPicture {
				t.Errorf("PictureURL: got %q, want %q", g.PictureURL, tc.expectedPicture)
			}
			if g.ParentName != tc.expectedParent {
				t.Errorf("ParentName: got %q, want %q", g.ParentName, tc.expectedParent)
			}
			if g.GrafanaDashboardSelector != tc.expectedSelector {
				t.Errorf("GrafanaDashboardSelector: got %q, want %q", g.GrafanaDashboardSelector, tc.expectedSelector)
			}
			if !reflect.DeepEqual(g.MemberNames, tc.expectedMembers) {
				t.Errorf("MemberNames: got %v, want %v", g.MemberNames, tc.expectedMembers)
			}
		})
	}
}
