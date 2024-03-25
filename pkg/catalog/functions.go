package catalog

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/giantswarm/backstage-catalog-importer/pkg/repositories"
)

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
	dependsOn []string) Entity {
	e := Entity{
		APIVersion: "backstage.io/v1alpha1",
		Kind:       EntityKindComponent,
		Metadata: EntityMetadata{
			Name:        r.Name,
			Labels:      map[string]string{},
			Description: description,
			Annotations: map[string]string{
				"github.com/project-slug":      fmt.Sprintf("giantswarm/%s", r.Name),
				"github.com/team-slug":         team,
				"backstage.io/source-location": fmt.Sprintf("url:https://github.com/giantswarm/%s", r.Name),
				"quay.io/repository-slug":      fmt.Sprintf("giantswarm/%s", r.Name),

				// Like team name, but without "team-" prefix
				"opsgenie.com/team": strings.TrimPrefix(team, "team-"),
			},
			Tags: []string{},
		},
	}

	// Additional metadata
	{
		if latestReleaseTag != "" {
			e.Metadata.Annotations["giantswarm.io/latest-release-tag"] = latestReleaseTag
			e.Metadata.Annotations["giantswarm.io/latest-release-date"] = latestReleaseTime.Format(time.RFC3339)
		}

		if hasCircleCi {
			e.Metadata.Annotations["circleci.com/project-slug"] = fmt.Sprintf("github/giantswarm/%s", r.Name)
		}

		if hasReadme && defaultBranch != "" {
			e.Metadata.Annotations["backstage.io/techdocs-ref"] = fmt.Sprintf("url:https://github.com/giantswarm/%s/tree/%s", r.Name, defaultBranch)
		}

		// Possible deployment names for resource discovery via the Giant Swarm plugin,
		// Grafana dashboards, and OpsGenie alerts.
		deploymentNames := r.DeploymentNames
		sort.Strings(deploymentNames)

		// Default deployment names are REPONAME and REPONAME-app.
		if len(deploymentNames) == 0 {
			name := strings.TrimSuffix(r.Name, "-app")
			nameWithAppSuffix := fmt.Sprintf("%s-app", name)
			deploymentNames = []string{
				name,
				nameWithAppSuffix,
			}
		}

		e.Metadata.Annotations["giantswarm.io/deployment-names"] = strings.Join(deploymentNames, ",")

		// OpsGenie query
		queryParts := []string{}
		for _, d := range deploymentNames {
			queryParts = append(queryParts, fmt.Sprintf("detailsPair(app:%s)", d))
		}
		e.Metadata.Annotations["opsgenie.com/component-selector"] = strings.Join(queryParts, " OR ")

		if r.Gen.Language != "" && r.Gen.Language != repositories.RepoLanguageGeneric {
			e.Metadata.Labels["giantswarm.io/language"] = string(r.Gen.Language)

			e.Metadata.Tags = append(e.Metadata.Tags, fmt.Sprintf("language:%s", r.Gen.Language))
		}

		if isPrivate {
			e.Metadata.Tags = append(e.Metadata.Tags, "private")
		}

		if defaultBranch == "master" {
			e.Metadata.Tags = append(e.Metadata.Tags, "defaultbranch:master")
		}

		if latestReleaseTag == "" {
			e.Metadata.Tags = append(e.Metadata.Tags, "no-releases")
		}

		for _, flavor := range r.Gen.Flavors {
			e.Metadata.Labels[fmt.Sprintf("giantswarm.io/flavor-%s", flavor)] = "true"

			e.Metadata.Tags = append(e.Metadata.Tags, fmt.Sprintf("flavor:%s", flavor))
		}

		sort.Strings(e.Metadata.Tags)

		if r.ComponentType == "service" {
			// Kubernetes plugin annotation
			e.Metadata.Annotations["backstage.io/kubernetes-id"] = r.Name

			// Grafana dashboard links
			urlParts := []string{}
			for _, d := range deploymentNames {
				urlParts = append(urlParts, fmt.Sprintf("var-app=%s", d))
			}
			e.Metadata.Links = []EntityLink{
				{
					URL:   fmt.Sprintf("https://giantswarm.grafana.net/d/eb617ba1-209a-4d57-9963-1af9a8ddc8d4/general-service-metrics?orgId=1&%s&from=now-24h&to=now", strings.Join(urlParts, "&")),
					Title: "General service metrics dashboard",
					Icon:  "dashboard",
					Type:  "grafana-dashboard",
				},
			}
		}
	}

	// Entity spec

	spec := ComponentSpec{
		Type:      "unspecified",
		Lifecycle: "production",
		Owner:     team,
		System:    system,
	}

	if r.ComponentType != "" {
		spec.Type = r.ComponentType
	}

	if r.Lifecycle != "production" && r.Lifecycle != "" {
		spec.Lifecycle = string(r.Lifecycle)
	}

	if len(dependsOn) > 0 {
		for _, d := range dependsOn {
			spec.DependsOn = append(spec.DependsOn, "component:"+d)
		}
	}

	e.Spec = spec

	return e
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
