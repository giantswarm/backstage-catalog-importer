# Backstage catalog importer documentation

Backstage catalog importer generates Backstage catalog YAML files from a variety of sources.

Sources are:

- **GitHub repositories**: these get exported as Component entities
- **GitHub teams**: these get exported as Group entities
- **GitHub users**: these get exported as User entities
- **Giant Swarm app catalogs**: these get exported as Component entities
- **Giant Swarm installations**: these get exported as Resource entities

Sources get discovered dynamically.

The tool covers two different use cases so far:

1. **Giant Swarm dev portal**: export entities internal to Giant Swarm (various components, teams, users)
2. **Customer catalogs**: export entities for a customer Backstage instance (components from app catalogs, and teams owning these components)

## GitHub repository discovery

The tool processes all repository lists in the private folder https://github.com/giantswarm/github/tree/main/repositories.

Repository lists are expected to provide some Backstage specific metadata, for example: the component type (service, library, ...), and the names to look up for finding related Kubernetes deployments.

TODO: This method is suited for the Giant Swarm dev portal only. For customers, we need a different solution. For example, Backstage catalog info YAML files could be placed in each relevant repository. Giant Swarm specific metadata would have to be added in order to unlock certain features.

## GitHub team discovery

To discover teams, we read all teams of the `giantswarm` organization from the GitHub API.

TODO: This method is suited for Giant Swarm only. For customer catalogs, we need a different solution. First, we don't want to export information on teams that are not relevant to the customer. Second, we may want to publish customer's teams.

## User discovery

We export all Personio employee entries that have the `status` field set to `active`.

This requires `PERSONIO_CLIENT_ID` and `PERSONIO_CLIENT_SECRET` to be configured for an API token with the following minimal scope:
- `dynamic_3196204` (employee GitHub handle)
- `first_name`
- `last_name`
- `email`
- `status`

## Giant Swarm app catalog discovery

App catalog URLs are hard-coded. All of them are publicly accessible.

TODO: We need flexibility to skip catalogs, and to add customer-specific catalogs (non-public), too.

## Giant Swarm installation discovery

Installations are read from a private GitHub repository. The resulting catalog is meant for use within Giant Swarm only.
