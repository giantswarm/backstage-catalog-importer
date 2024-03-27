module github.com/giantswarm/backstage-catalog-importer

go 1.21

toolchain go1.22.0

require (
	github.com/giantswarm/microerror v0.4.1
	github.com/google/go-cmp v0.6.0
	github.com/google/go-github/v60 v60.0.0
	github.com/spf13/cobra v1.8.0
	golang.org/x/exp v0.0.0-20240318143956-a85f2c67cd81
	golang.org/x/oauth2 v0.18.0
	gopkg.in/yaml.v3 v3.0.1
	helm.sh/helm/v3 v3.14.3
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace golang.org/x/net => golang.org/x/net v0.22.0
