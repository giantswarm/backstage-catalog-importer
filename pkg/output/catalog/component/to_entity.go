package component

import (
	"fmt"
	"sort"
	"strings"

	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/output/bscatalog/v1alpha1"
)

// Returns an entity representation of the Component.
func (c *Component) ToEntity() *bscatalog.Entity {
	tags := make([]string, len(c.Tags))
	copy(tags, c.Tags)
	if c.IsPrivate {
		tags = append(tags, "private")
	}
	if c.DefaultBranch == "master" {
		tags = append(tags, "defaultbranch:master")
	}
	for _, flavor := range c.Flavors {
		tags = append(tags, fmt.Sprintf("flavor:%s", flavor))
	}
	if len(c.HelmCharts) > 0 {
		tags = append(tags, "helmchart")
	}
	sort.Strings(tags)

	e := &bscatalog.Entity{
		APIVersion: bscatalog.APIVersion,
		Kind:       bscatalog.EntityKindComponent,
		Metadata: bscatalog.EntityMetadata{
			Annotations: make(map[string]string),
			Description: c.Description,
			Labels:      make(map[string]string),
			Links:       make([]bscatalog.EntityLink, 0),
			Name:        c.Name,
			Title:       c.Title,
		},
	}

	if c.Namespace != defaultNamespace {
		e.Metadata.Namespace = c.Namespace
	}

	if len(c.Annotations) > 0 {
		e.Metadata.Annotations = make(map[string]string)
		for k, v := range c.Annotations {
			e.Metadata.Annotations[k] = v
		}
	}

	if len(c.Labels) > 0 {
		e.Metadata.Labels = make(map[string]string)
		for k, v := range c.Labels {
			e.Metadata.Labels[k] = v
		}
	}

	if c.Language != "" {
		if e.Metadata.Labels == nil {
			e.Metadata.Labels = make(map[string]string)
		}
		e.Metadata.Labels["giantswarm.io/language"] = c.Language
	}

	if len(c.Flavors) > 0 {
		if e.Metadata.Labels == nil {
			e.Metadata.Labels = make(map[string]string)
		}
		for _, flavor := range c.Flavors {
			e.Metadata.Labels[fmt.Sprintf("giantswarm.io/flavor-%s", flavor)] = "true"
		}
	}

	if len(c.Links) > 0 {
		e.Metadata.Links = c.Links
	}

	if len(tags) > 0 {
		e.Metadata.Tags = tags
	}

	if c.Language != "" {
		e.Metadata.Tags = append(e.Metadata.Tags, fmt.Sprintf("language:%s", c.Language))
	}

	if c.GithubProjectSlug != "" {
		e.Metadata.Annotations["github.com/project-slug"] = c.GithubProjectSlug
		e.Metadata.Annotations["backstage.io/source-location"] = fmt.Sprintf("url:https://github.com/%s", c.GithubProjectSlug)
		if c.HasReadme && c.DefaultBranch != "" {
			e.Metadata.Annotations["backstage.io/techdocs-ref"] = fmt.Sprintf("url:https://github.com/%s/tree/%s", c.GithubProjectSlug, c.DefaultBranch)
		}
	}
	if c.GithubTeamSlug != "" {
		e.Metadata.Annotations["github.com/team-slug"] = c.GithubTeamSlug
	}
	if c.CircleCiSlug != "" {
		e.Metadata.Annotations["circleci.com/project-slug"] = c.CircleCiSlug
	}
	if c.Type == typeService {
		e.Metadata.Annotations["backstage.io/kubernetes-id"] = c.Name
		if c.KubernetesID != "" {
			e.Metadata.Annotations["backstage.io/kubernetes-id"] = c.KubernetesID
		}
	}
	if len(c.HelmCharts) > 0 {
		versions := make([]string, len(c.HelmCharts))
		appVersions := make([]string, len(c.HelmCharts))
		fullChartNames := make([]string, len(c.HelmCharts))
		for i, chart := range c.HelmCharts {
			fullChartNames[i] = fmt.Sprintf("%s/%s/%s", c.OciRegistry, c.OciRepositoryPrefix, chart.Name)

			// A version is withheld rather than removed: the frontend pairs
			// these comma-separated lists with helmcharts by index, so the
			// slots have to stay aligned.
			//
			// The app version is nested inside the chart-version check, as in
			// cmd/charts: both values are read from the same Chart.yaml, so an
			// unsubstituted chart version means the app version beside it
			// describes an unreleased state too and must not be published on
			// its own.
			if isPublishableChartVersion(chart.Version) {
				versions[i] = chart.Version

				if isPublishableAppVersion(chart.AppVersion) {
					appVersions[i] = chart.AppVersion
				}
			}
		}

		e.Metadata.Annotations["giantswarm.io/helmcharts"] = strings.Join(fullChartNames, ",")
		if hasAny(versions) {
			e.Metadata.Annotations["giantswarm.io/helmchart-versions"] = strings.Join(versions, ",")
		}
		if hasAny(appVersions) {
			e.Metadata.Annotations["giantswarm.io/helmchart-app-versions"] = strings.Join(appVersions, ",")
		}
	}

	spec := bscatalog.ComponentSpec{
		Type:      c.Type,
		Lifecycle: c.Lifecycle,
		Owner:     c.Owner,
	}
	if c.System != "" {
		spec.System = c.System
	}
	if len(c.DependsOn) > 0 {
		sort.Strings(c.DependsOn)
		for i, d := range c.DependsOn {
			c.DependsOn[i] = "component:" + d
		}
		spec.DependsOn = c.DependsOn
	}

	e.Metadata.NormalizeTags()
	e.Spec = spec

	return e
}
