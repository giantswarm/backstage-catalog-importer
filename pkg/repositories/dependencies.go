package repositories

import (
	"net/http"
	"regexp"

	"github.com/giantswarm/microerror"
	"golang.org/x/exp/slices"
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
func (s *Service) GetDependencies(name string) ([]string, error) {
	sbom, resp, err := s.githubClient.DependencyGraph.GetSBOM(s.ctx, s.config.GithubOrganization, name)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return nil, microerror.Mask(dependenciesNotFoundError)
		}
		return nil, err
	}

	if sbom == nil || sbom.SBOM == nil {
		return nil, microerror.Mask(dependenciesNotFoundError)
	}

	names := []string{}
	godepRegex := regexp.MustCompile("go:github.com/giantswarm/([^/]+).*")

	for _, item := range sbom.SBOM.Packages {
		// We only want these:
		// 'go:github.com/giantswarm/NAME'
		matches := godepRegex.FindStringSubmatch(*item.Name)
		if len(matches) > 0 {
			names = append(names, matches[1])
		}
	}

	slices.Sort(names)
	names = slices.Compact(names)

	return names, nil
}
