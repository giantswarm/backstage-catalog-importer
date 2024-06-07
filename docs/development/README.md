# Development docs

## Package overview

Our packages can be grouped into two main categories:

### Input source reading

- `repositories` - Provides means to read Giant Swarm's GitHub [repositories](https://github.com/giantswarm/github/tree/master/repositories) configuration, their GitHub API data, and some of their content from the GitHub API. This is highly Giant Swarm specific and must be untangled.
- `appcatalog` - Provides means to parse Giant Swarm app platform catalogs and read apps from them.
- `teams` - Provides means to read GitHub teams and their members from the GitHub API.
- `helmchart` - Simple helper to parse Helm chart YAML files published by Giant Swarm.

### Output generation

- `catalog` - High level objects, which are turned into bscatalog entities. These are the glue between input and output.
- `bscatalog` - Lower level Backstage catalog entities, which can be marshaled to YAML. There is a `v1alpha1` sub package for the current version of the entities API.
- `export` - Handles the creation of export YAML files.
- `legacy` - Functions that need a new home.
