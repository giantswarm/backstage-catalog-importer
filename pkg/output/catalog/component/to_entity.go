package component

import (
	"fmt"
	"sort"
	"strings"
	"time"

	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/output/bscatalog/v1alpha1"
)

// Returns an entity representation of the Component.
func (c *Component) ToEntity() *bscatalog.Entity {
	sort.Strings(c.Tags)

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

	if c.Namespace != "default" {
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

	if len(c.Links) > 0 {
		e.Metadata.Links = c.Links
	}

	if c.Tags != nil {
		e.Metadata.Tags = c.Tags
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
	if c.LatestReleaseTag != "" {
		e.Metadata.Annotations["giantswarm.io/latest-release-tag"] = c.LatestReleaseTag
	}
	if c.LatestReleaseTime.Format(time.RFC3339) != "0001-01-01T00:00:00Z" {
		e.Metadata.Annotations["giantswarm.io/latest-release-date"] = c.LatestReleaseTime.Format(time.RFC3339)
	}
	if c.CircleCiSlug != "" {
		e.Metadata.Annotations["circleci.com/project-slug"] = c.CircleCiSlug
	}
	if c.DeploymentNames != nil {
		sort.Strings(c.DeploymentNames)
		e.Metadata.Annotations["giantswarm.io/deployment-names"] = strings.Join(c.DeploymentNames, ",")
	}
	if c.Type == "service" {
		e.Metadata.Annotations["backstage.io/kubernetes-id"] = c.Name
		if c.KubernetesID != "" {
			e.Metadata.Annotations["backstage.io/kubernetes-id"] = c.KubernetesID
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
