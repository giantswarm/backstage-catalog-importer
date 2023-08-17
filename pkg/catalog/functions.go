package catalog

import (
	"fmt"
	"sort"

	"github.com/giantswarm/backstage-catalog-importer/pkg/repositories"
)

func CreateComponentEntity(r repositories.Repo, team, description string, system string, isPrivate bool, hasCircleCi, hasReadme bool, defaultBranch string, dependsOn []string) Entity {
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
			},
			Tags: []string{},
		},
	}

	if hasCircleCi {
		e.Metadata.Annotations["circleci.com/project-slug"] = fmt.Sprintf("github/giantswarm/%s", r.Name)
	}

	if hasReadme && defaultBranch != "" {
		e.Metadata.Annotations["backstage.io/techdocs-ref"] = fmt.Sprintf("url:https://github.com/giantswarm/%s/tree/%s", r.Name, defaultBranch)
	}

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

	if r.Gen.Language != "" && r.Gen.Language != repositories.RepoLanguageGeneric {
		e.Metadata.Labels["giantswarm.io/language"] = string(r.Gen.Language)

		e.Metadata.Tags = append(e.Metadata.Tags, fmt.Sprintf("language:%s", r.Gen.Language))
	}

	if isPrivate {
		e.Metadata.Tags = append(e.Metadata.Tags, "private")
	}

	for _, flavor := range r.Gen.Flavors {
		e.Metadata.Labels[fmt.Sprintf("giantswarm.io/flavor-%s", flavor)] = "true"

		e.Metadata.Tags = append(e.Metadata.Tags, fmt.Sprintf("flavor:%s", flavor))
	}

	if r.ComponentType == "service" {
		// Add Grafana dashboard links
		e.Metadata.Links = []EntityLink{
			{
				URL:   fmt.Sprintf("https://giantswarm.grafana.net/d/eb617ba1-209a-4d57-9963-1af9a8ddc8d4/general-service-metrics?orgId=1&var-app=%s&from=now-24h&to=now", r.Name),
				Title: "General service metrics dashboard",
				Icon:  "dashboard",
				Type:  "grafana-dashboard",
			},
		}
	}

	return e
}

func CreateGroupEntity(name, displayName, description, parent string, members []string, id int64) Entity {
	sort.Strings(members)
	e := Entity{
		APIVersion: "backstage.io/v1alpha1",
		Kind:       EntityKindGroup,
		Metadata: EntityMetadata{
			Name: name,
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
