module github.com/giantswarm/backstage-catalog-importer

go 1.22.0

toolchain go1.22.4

require (
	github.com/giantswarm/microerror v0.4.1
	github.com/google/go-cmp v0.6.0
	github.com/google/go-github/v62 v62.0.0
	github.com/spf13/cobra v1.8.1
	golang.org/x/exp v0.0.0-20240613232115-7f521ea00fb8
	golang.org/x/oauth2 v0.21.0
	gopkg.in/yaml.v3 v3.0.1
	helm.sh/helm/v3 v3.15.2
	sigs.k8s.io/yaml v1.4.0
)

require (
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace golang.org/x/net => golang.org/x/net v0.26.0
