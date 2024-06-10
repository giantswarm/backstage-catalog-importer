# Usage

## Generate catalog files for the Giant Swarm developer portal

This requires a Github personal access token (PTA) with permission to read repository content, teams, and user info for the `giantswarm` organization, provided as `GITHUB_TOKEN` environment variable.

To run the export, execute

```nohighlight
backstage-catalog-importer [--output path-to-output-dir]
```

As a result, several YAML files will be written to the output directory. Progress and warnings will be logged to the console.

### What's covered

The following data will be included in the generated catalog:

- All repositories referenced in the repositories lists in [giantswarm/github](https://github.com/giantswarm/github/tree/main/repositories) as _Component_ entities.
- All teams of the configured Github organizaiton as _Group_ entities.
- All members of the above teams as _User_ entities.
