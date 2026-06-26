package groups

import (
	"reflect"
	"testing"

	"github.com/google/go-github/v88/github"
)

func TestIsDescendant(t *testing.T) {
	// Hierarchy:
	//   employees
	//     ├── team-atlas
	//     │     └── team-atlas-sub
	//     └── ae-adidas
	//   bots            (no parent)
	parentBySlug := map[string]string{
		"employees":      "",
		"team-atlas":     "employees",
		"team-atlas-sub": "team-atlas",
		"ae-adidas":      "employees",
		"bots":           "",
	}

	testCases := []struct {
		slug     string
		ancestor string
		expected bool
	}{
		{slug: "team-atlas", ancestor: "employees", expected: true},       // direct child
		{slug: "team-atlas-sub", ancestor: "employees", expected: true},   // transitive descendant
		{slug: "ae-adidas", ancestor: "employees", expected: true},        // customer team is also under employees
		{slug: "employees", ancestor: "employees", expected: false},       // a team is not its own descendant
		{slug: "bots", ancestor: "employees", expected: false},            // unrelated top-level team
		{slug: "team-atlas", ancestor: "team-atlas-sub", expected: false}, // ancestor/descendant reversed
		{slug: "unknown", ancestor: "employees", expected: false},         // unknown slug
	}

	for _, tc := range testCases {
		t.Run(tc.slug+"_in_"+tc.ancestor, func(t *testing.T) {
			got := isDescendant(tc.slug, tc.ancestor, parentBySlug)
			if got != tc.expected {
				t.Errorf("isDescendant(%q, %q): got %v, want %v", tc.slug, tc.ancestor, got, tc.expected)
			}
		})
	}
}

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
