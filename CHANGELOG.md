# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Introduce new `pkg/catalog/Component` type as an abstraction for a component entity.
- Introduce new `pkg/catalog/group/Group` type as an abstraction for a group entity.
- Add command `appcatalogs` to export Giant Swarm app catalogs.

### Changed

- Function `pkg/catalog/CreateComponentEntity` is now deprecated. We want to use the new `Component` type and its ToEntity method instead.

### Fixed

- Tool name in `--help` output is now correct.
- The component annotation `giantswarm.io/latest-release-date` is no longer rendered if the value would be `0001-01-01T00:00:00Z`.

## [0.12.0] - 2024-03-28

### Added

- Add the tag `helmchart` if the component provides one or more Helm charts.
- Add annotations regarding Helm charts to exported component entities, in case the component provides one or several charts:
  - `giantswarm.io/helmchart-app-versions`: a list of upstream app versions provided by the component's helm charts, separated by comma.
  - `giantswarm.io/helmchart-versions`: a list of Helm chart versions provided by the component, separated by comma.
  - `giantswarm.io/helmcharts`: a list of names of the Helm charts provided by the component, separated by comma.

## [0.11.2] - 2024-03-25

### Changed

- Modify annotation keys:
  - `backstage.giantswarm.io/latest-release-tag` changed to `giantswarm.io/latest-release-tag`
  - `backstage.giantswarm.io/latest-release-date` changed to `giantswarm.io/latest-release-date`

## [0.11.1] - 2024-03-15

### Fixed

- Fixed nil pointer dereference problem

## [0.11.0] - 2024-03-15

### Added

- Add annotations about the latest release of a component entity:
  - `backstage.giantswarm.io/latest-release-tag`
  - `backstage.giantswarm.io/latest-release-date`
- Add tag `no-releases` in case there are no releases for a component
- Adds tag `defaultbranch:master` if the component repo uses master as the default branch name

### Removed

- Remove `dependabotRemove` key from repositories data processing

## [0.10.0] - 2024-02-26

### Added

- Add support for the `deploymentNames` field in repository metadata.

## [0.9.0] - 2024-02-07

### Added

- For components of type "service", we add kubernetes-id annotation.

## [0.8.0] - 2024-01-16

### Added

- For components of type "service", we add a list of possible deployment names.

## [0.7.0] - 2023-09-22

### Changed

- Export one file per entity kind instead of one common file.

### Removed

- Removed `--format` flag and ability to export a Kubernetes ConfigMap.

## [0.6.0] - 2023-09-19

### Added

- Add reporting for Go dependencies that are used by other components in the catalog, but not imported into the catalog. This appears in the end of every workflow run log.

## [0.5.0] - 2023-09-14

- Add OpsGenie related entity annotations.

## [0.4.0] - 2023-08-22

### Removed

- Remove catalog data export from static YAML files.

## [0.3.1] - 2023-08-21

- Fix dashboard selector.

## [0.3.0] - 2023-08-21

- Add `grafana/dashboard-selector` annotation to Group entities, to enable showing of dashboards for teams.

## [0.2.1] - 2023-08-18

- Ensure deterministic order of dependencies, avoid duplicates.
- Simplify Github dependency graph query.

## [0.2.0] - 2023-08-17

### Added

- For components of type "service", we add a Grafana dashboard link.
- Export catalog data from static YAML files.
- Accept `system` property when reading repository data.

### Removed

- Removed debug logging on dependencies found.

## [0.1.0] - 2023-08-16

### Added

- Export `dependsOn` relationships between a Go component and the libraries it is using (if the library is ours).

## [0.0.9] - 2023-08-08

### Added

- Set a component type based on the data read from giantswarm/github.

## [0.0.8] - 2023-08-07

### Fixed

- Accept `componentType` property when reading repository data.

## [0.0.7] - 2023-08-02

### Added

- Add `backstage.io/techdocs-ref` annotation for components (only if `/README.md` is present).

### Changed

- Only create `circleci.com/project-slug` annotation on component if there is a `.circleci/config.yml` file.

## [0.0.6] - 2023-07-27

### Added

- Export repository descriptions to component entities.
- Export repository visibility (private) to component entity tags.

## [0.0.5] - 2023-07-26

### Changed

- Sort Group members alphanumerically.

## [0.0.4] - 2023-07-26

- Treat all GitHub teams in the organization, instead of only the teams owning repositories.

## [0.0.3] - 2023-07-25

### Fixed

- Write proper ConfigMap manifest.

## [0.0.2] - 2023-07-24

### Added

- Added option to write ConfigMap output.

## [0.0.1] - 2023-07-24

### Added

- Initial code

[Unreleased]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.12.0...HEAD
[0.12.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.11.2...v0.12.0
[0.11.2]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.11.1...v0.11.2
[0.11.1]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.11.0...v0.11.1
[0.11.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.10.0...v0.11.0
[0.10.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.9.0...v0.10.0
[0.9.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.0.9...v0.1.0
[0.0.9]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.0.8...v0.0.9
[0.0.8]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.0.7...v0.0.8
[0.0.7]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.0.6...v0.0.7
[0.0.6]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.0.5...v0.0.6
[0.0.5]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.0.4...v0.0.5
[0.0.4]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.0.3...v0.0.4
[0.0.3]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/giantswarm/backstage-catalog-importer/releases/tag/v0.0.1
