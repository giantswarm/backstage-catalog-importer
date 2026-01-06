package component

import (
	"time"

	"github.com/giantswarm/backstage-catalog-importer/pkg/input/helmchart"
)

// Option is an option to configure a Component.
type Option func(*Component)

func WithNamespace(namespace string) Option {
	return func(c *Component) {
		c.Namespace = namespace
	}
}

func WithDescription(description string) Option {
	return func(c *Component) {
		c.Description = description
	}
}

func WithTitle(title string) Option {
	return func(c *Component) {
		c.Title = title
	}
}

func WithGithubTeamSlug(team string) Option {
	return func(c *Component) {
		c.GithubTeamSlug = team
	}
}

func WithSystem(system string) Option {
	return func(c *Component) {
		c.System = system
	}
}

func WithCircleCiSlug(slug string) Option {
	return func(c *Component) {
		c.CircleCiSlug = slug
	}
}

func WithHasReadme(hasReadme bool) Option {
	return func(c *Component) {
		c.HasReadme = hasReadme
	}
}

func WithDefaultBranch(defaultBranch string) Option {
	return func(c *Component) {
		c.DefaultBranch = defaultBranch
	}
}

func WithLatestReleaseTime(latestReleaseTime time.Time) Option {
	return func(c *Component) {
		c.LatestReleaseTime = latestReleaseTime
	}
}

func WithLatestReleaseTag(latestReleaseTag string) Option {
	return func(c *Component) {
		c.LatestReleaseTag = latestReleaseTag
	}
}

func WithHasReleases(hasReleases bool) Option {
	return func(c *Component) {
		c.HasReleases = hasReleases
	}
}

func WithGithubProjectSlug(slug string) Option {
	return func(c *Component) {
		c.GithubProjectSlug = slug
	}
}

func WithDeploymentNames(names ...string) Option {
	return func(c *Component) {
		c.DeploymentNames = names
	}
}

func WithType(t string) Option {
	return func(c *Component) {
		c.Type = t
	}
}

func WithLifecycle(lifecycle string) Option {
	return func(c *Component) {
		if lifecycle != "" {
			c.Lifecycle = lifecycle
		}
	}
}

func WithDependsOn(dependsOn ...string) Option {
	return func(c *Component) {
		c.DependsOn = dependsOn
	}
}

func WithLanguage(language string) Option {
	return func(c *Component) {
		c.Language = language
	}
}

func WithPrivate(isPrivate bool) Option {
	return func(c *Component) {
		c.IsPrivate = isPrivate
	}
}

func WithFlavors(flavors ...string) Option {
	return func(c *Component) {
		c.Flavors = flavors
	}
}

func WithTags(tags ...string) Option {
	return func(c *Component) {
		c.Tags = tags
	}
}

func WithLabels(labels map[string]string) Option {
	return func(c *Component) {
		c.Labels = labels
	}
}

func WithKubernetesID(id string) Option {
	return func(c *Component) {
		c.KubernetesID = id
	}
}

func WithOwner(owner string) Option {
	return func(c *Component) {
		c.Owner = owner
	}
}

func WithHelmCharts(charts ...*helmchart.Chart) Option {
	return func(c *Component) {
		c.HelmCharts = charts
	}
}

func WithOciRegistry(registry string) Option {
	return func(c *Component) {
		c.OciRegistry = registry
	}
}

func WithOciRepositoryPrefix(prefix string) Option {
	return func(c *Component) {
		c.OciRepositoryPrefix = prefix
	}
}
