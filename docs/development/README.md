# Development

## Package overview

Our packages can be grouped into two main categories:

### Input source reading

- `input/repositories` - Provides means to read Giant Swarm's GitHub [repositories](https://github.com/giantswarm/github/tree/master/repositories) configuration, their GitHub API data, and some of their content from the GitHub API. This is highly Giant Swarm specific and must be untangled.
- `input/appcatalog` - Provides means to parse Giant Swarm app platform catalogs and read apps from them.
- `input/teams` - Provides means to read GitHub teams and their members from the GitHub API.
- `input/helmchart` - Simple helper to parse Helm chart YAML files published by Giant Swarm.

### Output generation

- `output/catalog` - High level objects, which are turned into `bscatalog` entities, but also support well-known annotations and labels. These are the glue between input and output.
- `output/bscatalog` - Lower level Backstage catalog entities, which can be marshaled to YAML. There is a `v1alpha1` sub package for the current version of the entities API.
- `output/export` - Handles the creation of export YAML files from `bscatalog.Entity` objects.
- `output/legacy` - Functions that need a new home.
