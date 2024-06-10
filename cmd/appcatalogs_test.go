package cmd

import (
	"testing"

	"github.com/giantswarm/backstage-catalog-importer/pkg/input/helmrepoindex"
)

func Test_githubSlugFromURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "without-trailing-slash",
			url:  "https://github.com/giantswarm/app-catalog",
			want: "giantswarm/app-catalog",
		},
		{
			name: "with-trailing-slash",
			url:  "https://github.com/giantswarm/app-catalog/",
			want: "giantswarm/app-catalog",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := githubSlugFromURL(tt.url); got != tt.want {
				t.Errorf("githubSlugFromURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_detectGitHubSlug(t *testing.T) {
	tests := []struct {
		name  string
		entry *helmrepoindex.Entry
		want  string
	}{
		{
			name: "nothing-found",
			entry: &helmrepoindex.Entry{
				Home: "https://www.giantswarm.io/",
				Sources: []string{
					"https://codeberg.org/foo/bar",
				},
				Urls: []string{
					"https://www.giantswarm.io/",
				},
			},
			want: "",
		},
		{
			name: "url-found-in-home",
			entry: &helmrepoindex.Entry{
				Home: "https://github.com/giantswarm/project-slug",
				Sources: []string{
					"https://codeberg.org/foo/bar",
				},
				Urls: []string{
					"https://www.giantswarm.io/",
				},
			},
			want: "giantswarm/project-slug",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectGitHubSlug(tt.entry); got != tt.want {
				t.Errorf("detectGitHubSlug() = %v, want %v", got, tt.want)
			}
		})
	}
}
