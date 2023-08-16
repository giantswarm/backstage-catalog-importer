package repositories

import (
	"fmt"
	"regexp"

	"github.com/google/go-github/v53/github"
)

type SbomPayload struct {
	Sbom Sbom `json:"sbom"`
}

type Sbom struct {
	Name     string        `json:"name"`
	Packages []SbomPackage `json:"packages"`
}

type SbomPackage struct {
	Name string `json:"name"`
}

// Returns list of dependencies.
//
// As long as https://github.com/google/go-github/issues/2864 is open,
// we have to make this a low-level request.
func (s *Service) GetDependencies(name string) ([]string, error) {
	path := fmt.Sprintf("/repos/%s/%s/dependency-graph/sbom", s.config.GithubOrganization, name)
	req, err := s.githubClient.NewRequest("GET", path, nil, github.WithVersion("2022-11-28"))
	if err != nil {
		return nil, err
	}

	names := []string{}
	payload := SbomPayload{}
	resp, err := s.githubClient.Do(s.ctx, req, &payload)
	if err != nil {
		if resp.StatusCode == 404 {
			return nil, dependenciesNotFoundError
		}
		return nil, err
	}

	godepRegex := regexp.MustCompile("go:github.com/giantswarm/([^/]+).*")

	for _, item := range payload.Sbom.Packages {
		// We only want these:
		// 'go:github.com/giantswarm/NAME'
		matches := godepRegex.FindStringSubmatch(item.Name)
		if len(matches) > 0 {
			names = append(names, matches[1])
		}
	}

	return uniq(names), nil
}

// Return de-duplicated strings slice.
func uniq(strings []string) []string {
	m := map[string]bool{}
	for _, item := range strings {
		m[item] = true
	}

	s := []string{}
	for key := range m {
		s = append(s, key)
	}

	return s
}
