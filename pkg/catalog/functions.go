package catalog

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/giantswarm/backstage-catalog-importer/pkg/helmchart"
	"github.com/giantswarm/backstage-catalog-importer/pkg/repositories"
)

// CreateComponentEntity is the deprecated way of generating a component entity.
// It sets some vaules in a Giant Swarm specific way. We are replacing it, with
// the goal to make this tool useful for more than just Giant Swarm.
// Deprecated: Use the Component struct and its ToEntity() method instead.
func CreateComponentEntity(r repositories.Repo,
	team string,
	description string,
	system string,
	isPrivate bool,
	hasCircleCi,
	hasReadme bool,
	defaultBranch string,
	latestReleaseTime time.Time,
	latestReleaseTag string,
	charts []*helmchart.Chart,
	dependsOn []string) Entity {

	// Possible deployment names for resource discovery via the Giant Swarm plugin,
	// Grafana dashboards, and OpsGenie alerts.
	deploymentNames := r.DeploymentNames

	// Default deployment names are REPONAME and REPONAME-app.
	if len(deploymentNames) == 0 {
		name := strings.TrimSuffix(r.Name, "-app")
		nameWithAppSuffix := fmt.Sprintf("%s-app", name)
		deploymentNames = []string{
			name,
			nameWithAppSuffix,
		}
	}

	// OpsGenie query
	opsGenieQueryParts := []string{}
	for _, d := range deploymentNames {
		opsGenieQueryParts = append(opsGenieQueryParts, fmt.Sprintf("detailsPair(app:%s)", d))
	}
	opsGenieQuery := strings.Join(opsGenieQueryParts, " OR ")

	c, err := NewComponent(r.Name,
		WithDescription(description),
		WithGithubProjectSlug(fmt.Sprintf("giantswarm/%s", r.Name)),
		WithGithubTeamSlug(team),
		WithOpsGenieTeam(strings.TrimPrefix(team, "team-")),
		WithQuayRepositorySlug(fmt.Sprintf("giantswarm/%s", r.Name)),
		WithLatestReleaseTag(latestReleaseTag),
		WithLatestReleaseTime(latestReleaseTime),
		WithCircleCiSlug(fmt.Sprintf("github/giantswarm/%s", r.Name)),
		WithDefaultBranch(defaultBranch),
		WithHasReadme(hasReadme),
		WithDeploymentNames(deploymentNames...),
		WithOpsGenieComponentSelector(opsGenieQuery),
		WithSystem(system),
		WithType(r.ComponentType),
		WithOwner(team),
		WithDependsOn(dependsOn...),
	)
	if err != nil {
		log.Fatalf("Could not create component: %s", err)
	}

	// Additional metadata

	if r.Gen.Language != "" && r.Gen.Language != repositories.RepoLanguageGeneric {
		c.SetLabel("giantswarm.io/language", string(r.Gen.Language))
		c.AddTag(fmt.Sprintf("language:%s", r.Gen.Language))
	}

	if isPrivate {
		c.AddTag("private")
	}

	if defaultBranch == "master" {
		c.AddTag("defaultbranch:master")
	}

	if latestReleaseTag == "" {
		c.AddTag("no-releases")
	}

	if len(charts) > 0 {
		c.AddTag("helmchart")

		names := []string{}
		versions := []string{}
		appVersions := []string{}
		for _, c := range charts {
			names = append(names, c.Metadata.Name)
			versions = append(versions, c.Metadata.Version)
			appVersions = append(appVersions, c.Metadata.AppVersion)
		}

		c.SetAnnotation("giantswarm.io/helmcharts", strings.Join(names, ","))
		c.SetAnnotation("giantswarm.io/helmchart-versions", strings.Join(versions, ","))
		c.SetAnnotation("giantswarm.io/helmchart-app-versions", strings.Join(appVersions, ","))
	}

	for _, flavor := range r.Gen.Flavors {
		c.SetLabel(fmt.Sprintf("giantswarm.io/flavor-%s", flavor), "true")
		c.AddTag(fmt.Sprintf("flavor:%s", flavor))
	}

	// Grafana dashboard link
	if r.ComponentType == "service" {
		urlParts := []string{}
		for _, d := range deploymentNames {
			urlParts = append(urlParts, fmt.Sprintf("var-app=%s", d))
		}
		c.AddLink(EntityLink{
			URL:   fmt.Sprintf("https://giantswarm.grafana.net/d/eb617ba1-209a-4d57-9963-1af9a8ddc8d4/general-service-metrics?orgId=1&%s&from=now-24h&to=now", strings.Join(urlParts, "&")),
			Title: "General service metrics dashboard",
			Icon:  "dashboard",
			Type:  "grafana-dashboard",
		})
	}

	e := c.ToEntity()

	return *e
}

func CreateGroupEntity(name, displayName, description, parent string, members []string, id int64) Entity {
	sort.Strings(members)
	e := Entity{
		APIVersion: "backstage.io/v1alpha1",
		Kind:       EntityKindGroup,
		Metadata: EntityMetadata{
			Name: name,
			Annotations: map[string]string{
				"grafana/dashboard-selector": fmt.Sprintf("tags @> 'owner:%s'", name),

				// Like group name, but without "team-" prefix
				"opsgenie.com/team": strings.TrimPrefix(name, "team-"),
			},
		},
	}
	spec := GroupSpec{
		Type:    "team",
		Members: members,
		Profile: GroupProfile{
			Picture: fmt.Sprintf("https://avatars.githubusercontent.com/t/%d?s=116&v=4", id),
		},
	}

	if description != "" {
		e.Metadata.Description = description
	}
	if displayName != "" {
		spec.Profile.DisplayName = displayName
	}
	if parent != "" {
		spec.Parent = parent
	}

	e.Spec = spec

	return e
}

func CreateUserEntity(name, email, displayName, description, avatarURL string) Entity {
	e := Entity{
		APIVersion: "backstage.io/v1alpha1",
		Kind:       EntityKindUser,
		Metadata: EntityMetadata{
			Name: name,
		},
	}

	spec := UserSpec{
		MemberOf: []string{},
		Profile: UserProfile{
			Email: email,
		},
	}

	if description != "" {
		e.Metadata.Description = description
	}
	if displayName != "" {
		spec.Profile.DisplayName = displayName
	}
	if avatarURL != "" {
		spec.Profile.Picture = avatarURL
	}

	e.Spec = spec

	return e
}
