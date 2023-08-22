# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/giantswarm/backstage-catalog-importer/compare/v0.4.0...HEAD
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
