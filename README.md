# Backstage catalog importer

A utility to fill the catalog in the Giant Swarm developer portal.

## Usage

Requires a Github personal access token (PTA) with permission to read repository content, teams, and user info for the `giantswarm` organization, provided as `GITHUB_TOKEN` environment variable.

To run the export, execute

```nohighlight
go run main.go [--output output.yaml]
```

As a result, the file `output.yaml` in the current directory will provide the catalog content.

## What's covered

The following data will be exported/imported into the catalog:

- All repositories referenced in the repositories lists in [giantswarm/github](https://github.com/giantswarm/github/tree/main/repositories) as _Component_ entities.
- All teams of the configured Github organizaiton as _Group_ entities.
- All members of the above teams as _User_ entities.
