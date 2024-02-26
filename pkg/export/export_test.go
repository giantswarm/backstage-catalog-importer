package export

import (
	"flag"
	"io"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/backstage-catalog-importer/pkg/catalog"
	"github.com/giantswarm/backstage-catalog-importer/pkg/repositories"
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
		entities   []catalog.Entity
	}{
		{
			name:       "Group, users, and component",
			goldenFile: "case01.golden",
			entities: []catalog.Entity{
				catalog.CreateGroupEntity(
					"myorg/team-slug",
					"team-name",
					"A simple team with simple people",
					"area-everything",
					[]string{"jane-doe", "second-member"},
					16638849),
				catalog.CreateUserEntity(
					"jane-doe",
					"jane@acme.org",
					"Jane Doe",
					"Experienced DevOps engineer, jack of all trades",
					"https://avatars.githubusercontent.com/u/12345678?v=4"),
				catalog.CreateComponentEntity(
					repositories.Repo{
						Name:          "my-service",
						ComponentType: "service",
						System:        "everything-system",
						Gen: repositories.RepoGen{
							Flavors:                 []repositories.RepoFlavor{"app"},
							Language:                "go",
							InstallUpdateChart:      false,
							EnableFloatingMajorTags: false,
						},
						Lifecycle: "production",
						Replacements: repositories.RepoReplacements{
							ArchitectOrb:     true,
							Renovate:         true,
							PreCommit:        true,
							DependabotRemove: true,
						},
						AppTestSuite: t,
					},
					"myorg/team-slug",
					"Awesome microservice",
					"everything-system",
					false,
					true,
					true,
					"main",
					[]string{"first-dependency", "second-dependency"}),
			},
		},
		{
			name:       "Component with individual deployment names",
			goldenFile: "case02.golden",
			entities: []catalog.Entity{
				// catalog.CreateGroupEntity(
				// 	"myorg/team-slug",
				// 	"team-name",
				// 	"A simple team with simple people",
				// 	"area-everything",
				// 	[]string{"jane-doe", "second-member"},
				// 	16638849),
				// catalog.CreateUserEntity(
				// 	"jane-doe",
				// 	"jane@acme.org",
				// 	"Jane Doe",
				// 	"Experienced DevOps engineer, jack of all trades",
				// 	"https://avatars.githubusercontent.com/u/12345678?v=4"),
				catalog.CreateComponentEntity(
					repositories.Repo{
						Name:            "project-with-two-apps",
						ComponentType:   "service",
						DeploymentNames: []string{"first-name", "second-name-app"},
						System:          "everything-system",
						Gen: repositories.RepoGen{
							Flavors:  []repositories.RepoFlavor{"app", "generic"},
							Language: "go",
						},
						Lifecycle:    "production",
						Replacements: repositories.RepoReplacements{},
						AppTestSuite: t,
					},
					"myorg/team-slug",
					"Project that includes two apps",
					"everything-system",
					false,
					true,
					true,
					"main",
					[]string{"first-dependency", "second-dependency"}),
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
