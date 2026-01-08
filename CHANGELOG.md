# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.24.1] - 2026-01-08

### Fixed

- `charts`: Remove trailing slash from some names

## [0.24.0] - 2026-01-08

### Added

- Add `charts` command to generate a catalog based on an OCI registry.

## [0.23.1] - 2025-12-15

### Fixed

- Fix a nil pointer exception

## [0.23.0] - 2025-12-15

### Changed

- Change `giantswarm.io/helmcharts` entity annotation for components, to include registry hostname and path, like `gsoci.azurecr.io/charts/giantswarm/hello-world`.

### Added

- Add tag `helmchart-deployable` to component entity if component has a Helm chart of type application.

## [0.22.0] - 2025-09-10

### Removed

- Remove creation of the `quay.io/repository-slug` annotation.

## [0.21.2] - 2025-04-17

### Fixed

- The `appcatalogs` command now uses the repository slug found in chart URLs to find the repository for a chart.

## [0.21.1] - 2025-04-09

### Fixed

- Fix Policy Reporter URL in installation resources.

## [0.21.0] - 2025-04-08

### Added

- In installation resource entities, more links are added. One additional Grafana link and one to Policy Reporter, both via Teleport.

## [0.20.0] - 2025-02-25

### Added

- `installations` export now sets the `giantswarm.io/access-docs-markdown` annotation if extra access docs are provided in `docs/access.md` in the installation repository.

## [0.19.0] - 2025-01-27

### Added

- `installations` command sets the `giantswarm.io/custom-ca` annotation if the installation data provides a custom CA certificate.

### Removed

- The unused but exported function `GetInstallationFile` in `pkg/input/installations`.

## [0.18.1] - 2025-01-22

- Dependency update

## [0.18.0] - 2024-12-16

### Added

- Tags of created entities are normalized to some degree, to be compatible with Backstage requirements.

## [0.17.0] - 2024-12-12

### Added

- User export takes employee data from Personio into account

### Removed

- Removed `--org` flag from command `users`.

## [0.16.1] - 2024-11-22

### Fixed

- Fixed detection of Go dependencies via the Github API

## [0.16.0] - 2024-09-25

### Removed

- Remove Opsgenie related annotations

## [0.15.5] - 2024-09-17

### Fixed

- Fixed a nil-pointer dereference in the `installations` command.

## [0.15.4] - 2024-09-17

### Changed

- For installation resources, also read the region jey from the top level of the source file.

## [0.15.3] - 2024-08-12

### Added

- `installations` export now exports OpsGenie annotation to find alerts per installation.

## [0.15.2] - 2024-08-09

### Changed

- In the `installations` command, set the links for Happa and Grafana different for Vintage and CAPI.

## [0.15.1] - 2024-08-08

### Added

- The `installations` command also exports the base domain and account engineer name.

## [0.15.0] - 2024-08-03

### Added

- Add `installations` command to export entities for Giant Swarm installations.

## [0.14.0] - 2024-07-10

### Added

- Add deployment names to customer component entities.
- Add techdocs-ref annotation to customer component entities.

### Changed

- Make sure that exported entities are always sorted by API version, kind, namespace, and name.

## [0.13.3] - 2024-06-24

### Fixed

- Prevent setting component lifecycle to an empty value.

## [0.13.2] - 2024-06-21

### Fixed

- Fixed a bug where all components had lifecycle `production`.

### Changed

- Refactoring
  - Deleted `legacy.CreateGroupEntity`
  - Renamed `pkg.catalog.group.NewGroup` to `pkg.catalog.group.New`.
  - Moved `appcatalogs` command into `cmd/appcatalogs` package.

## [0.13.1] - 2024-06-20

### Changed

- Ensure stable sorting of `users` output by entity name, to ensure the same sorting as in v0.12.0 and earlier, when exported via the root command.
- Only export a user's `.metadata.title` if the Title attribute is set. Only export `.spec.profile.displayName` if the DisplayName attribute is set.
- Do not export the `.metadata.namespace` field for user entities if it is `"default"`.

### Removed

## [0.13.0] - 2024-06-20

**Breaking:** User catalog export has moved out of the root command.

### Added

- Add command `appcatalogs` to export Giant Swarm app catalogs.
- Add command `users` to export Giant Swarm people.
- Introduce new `pkg/catalog/component/Component` type as an abstraction for a component entity.
- Introduce new `pkg/catalog/group/Group` type as an abstraction for a group entity.
- Introduce new `pkg/catalog/user/User` type as an abstraction for a user entity.

### Changed

- The component annotation `giantswarm.io/latest-release-date` is no longer rendered if the value would be `0001-01-01T00:00:00Z`.

### Fixed

- Tool name in `--help` output is now correct.

### Removed

- Moved the creation of a user catalog out of the root command (see `users` command instead).

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

[Unreleased]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.24.1...HEAD
[0.24.1]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.24.0...v0.24.1
[0.24.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.23.1...v0.24.0
[0.23.1]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.23.0...v0.23.1
[0.23.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.22.0...v0.23.0
[0.22.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.21.2...v0.22.0
[0.21.2]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.21.1...v0.21.2
[0.21.1]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.21.0...v0.21.1
[0.21.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.20.0...v0.21.0
[0.20.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.19.0...v0.20.0
[0.19.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.18.1...v0.19.0
[0.18.1]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.18.0...v0.18.1
[0.18.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.17.0...v0.18.0
[0.17.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.16.1...v0.17.0
[0.16.1]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.16.0...v0.16.1
[0.16.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.15.5...v0.16.0
[0.15.5]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.15.4...v0.15.5
[0.15.4]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.15.3...v0.15.4
[0.15.3]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.15.2...v0.15.3
[0.15.2]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.15.1...v0.15.2
[0.15.1]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.15.0...v0.15.1
[0.15.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.14.0...v0.15.0
[0.14.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.13.3...v0.14.0
[0.13.3]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.13.2...v0.13.3
[0.13.2]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.13.1...v0.13.2
[0.13.1]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.13.0...v0.13.1
[0.13.0]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.12.0...v0.13.0
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
