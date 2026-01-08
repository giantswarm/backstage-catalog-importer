# TODO list for charts catalog

## Background

In general, we want to mostly align the charts metadata (Annotations and tags) with the components metadata.

Charts:

   backstage-catalog-importer charts <registry> [flags]

Components:

   backstage-catalog-importer [flags]

## Tasks

1. [x] Use GitHub repository name instead of chart name for Component .metadata.name
2. [x] Add `application.giantswarm.io/audience` annotation from Chart config annotations
3. [x] Add `application.giantswarm.io/managed` annotation from Chart config annotations
4. [x] Remove annotations `giantswarm.io/oci-registry` and `giantswarm.io/oci-repository`. Instead specify `giantswarm.io/helmcharts`
5. [x] Remove annotation `giantswarm.io/oci-tag`. Instead specify `giantswarm.io/helmchart-versions`.
6. [x] Add annotation `giantswarm.io/helmchart-app-versions` (see components)
7. [x] Add `giantswarm.io/deployment-names` annotation
8. [x] Remove the tag `helm-chart` and `oci` and add the tag `helmchart`.
9. [x] If the chart is deployable (see conditions for component catalog), add the tag `helmchart-deployable`.
