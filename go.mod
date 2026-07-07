module github.com/giantswarm/backstage-catalog-importer

go 1.26.0

toolchain go1.26.5

require (
	github.com/Masterminds/semver/v3 v3.5.0
	github.com/giantswarm/microerror v0.4.1
	github.com/giantswarm/personio-go v0.6.0
	github.com/google/go-cmp v0.7.0
	github.com/google/go-github/v89 v89.0.0
	github.com/opencontainers/image-spec v1.1.1
	github.com/spf13/cobra v1.10.2
	go.yaml.in/yaml/v3 v3.0.4
	helm.sh/helm/v3 v3.21.2
	oras.land/oras-go/v2 v2.6.1
	sigs.k8s.io/yaml v1.6.0
)

require (
	github.com/google/go-querystring v1.2.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	golang.org/x/sync v0.21.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace github.com/distribution/distribution/v3 v3.0.0 => github.com/distribution/distribution/v3 v3.1.1

replace go.opentelemetry.io/otel v1.43.0 => go.opentelemetry.io/otel v1.44.0

replace golang.org/x/crypto v0.49.0 => golang.org/x/crypto v0.53.0

replace golang.org/x/net v0.52.0 => golang.org/x/net v0.56.0

replace golang.org/x/sys v0.42.0 => golang.org/x/sys v0.46.0

replace github.com/containerd/containerd => github.com/containerd/containerd v1.7.33
