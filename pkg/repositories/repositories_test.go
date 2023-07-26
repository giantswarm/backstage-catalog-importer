// Package repositories provides types and tools to deal with
// Giant Swarm's repository configuration data maintained in
// https://github.com/giantswarm/github/tree/master/repositories
package repositories

import (
	"testing"

	cmp "github.com/google/go-cmp/cmp"
)

func TestLoadListShallow(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "Team Atlas",
			path:    "testdata/team-atlas.yaml",
			wantErr: false,
		},
		{
			name:    "Team Bigmac",
			path:    "testdata/team-bigmac.yaml",
			wantErr: false,
		},
		{
			name:    "Team Cabbage",
			path:    "testdata/team-cabbage.yaml",
			wantErr: false,
		},
		{
			name:    "Team Clippy",
			path:    "testdata/team-clippy.yaml",
			wantErr: false,
		},
		{
			name:    "Team Honey Badger",
			path:    "testdata/team-honeybadger.yaml",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := New(Config{
				GithubOrganization:   "foo",
				GithubRepositoryName: "bar",
				DirectoryPath:        "",
			})
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}
			_, err = s.LoadList(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestLoadList(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    []Repo
		wantErr bool
	}{
		{
			name: "Deep check of the artifical test case",
			args: args{path: "testdata/artificial.yaml"},
			want: []Repo{
				{
					Name: "name-only",
				},
				{
					Name: "repo-with-gen-part",
					Gen: RepoGen{
						Flavors:  []RepoFlavor{RepoFlavorApp},
						Language: RepoLanguageGeneric,
					},
				},
				{
					Name: "repo-with-replace-part",
					Replacements: RepoReplacements{
						ArchitectOrb: true,
						Renovate:     true,
					},
				},
				{
					Name: "generic-go",
					Gen: RepoGen{
						Flavors:                 []RepoFlavor{RepoFlavorGeneric},
						Language:                RepoLanguageGo,
						EnableFloatingMajorTags: true,
					},
					Lifecycle: "deprecated",
					Replacements: RepoReplacements{
						ArchitectOrb: false,
						PreCommit:    true,
						Renovate:     true,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := New(Config{
				GithubOrganization:   "foo",
				GithubRepositoryName: "bar",
				DirectoryPath:        "",
			})
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}

			got, err := s.LoadList(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("LoadList() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
