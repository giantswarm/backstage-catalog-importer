package githuburl

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseGitHubURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
		wantRef   string
		wantPath  string
		wantErr   bool
	}{
		{
			name:      "BlobURL",
			url:       "https://github.com/giantswarm/apiextensions/blob/main/config/crd/bases/apps.yaml",
			wantOwner: "giantswarm",
			wantRepo:  "apiextensions",
			wantRef:   "main",
			wantPath:  "config/crd/bases/apps.yaml",
			wantErr:   false,
		},
		{
			name:      "BlobURLWithBranch",
			url:       "https://github.com/org/repo/blob/feature/my-feature/path/to/file.yaml",
			wantOwner: "org",
			wantRepo:  "repo",
			wantRef:   "feature",
			wantPath:  "my-feature/path/to/file.yaml",
			wantErr:   false,
		},
		{
			name:      "BlobURLWithTag",
			url:       "https://github.com/kubernetes/kubernetes/blob/v1.28.0/api/openapi-spec/swagger.json",
			wantOwner: "kubernetes",
			wantRepo:  "kubernetes",
			wantRef:   "v1.28.0",
			wantPath:  "api/openapi-spec/swagger.json",
			wantErr:   false,
		},
		{
			name:      "RawURL",
			url:       "https://raw.githubusercontent.com/giantswarm/apiextensions/main/config/crd/bases/apps.yaml",
			wantOwner: "giantswarm",
			wantRepo:  "apiextensions",
			wantRef:   "main",
			wantPath:  "config/crd/bases/apps.yaml",
			wantErr:   false,
		},
		{
			name:      "RawURLWithCommitSHA",
			url:       "https://raw.githubusercontent.com/org/repo/abc123def456/path/file.yaml",
			wantOwner: "org",
			wantRepo:  "repo",
			wantRef:   "abc123def456",
			wantPath:  "path/file.yaml",
			wantErr:   false,
		},
		{
			name:      "HTTPWithoutS",
			url:       "http://github.com/org/repo/blob/main/file.yaml",
			wantOwner: "org",
			wantRepo:  "repo",
			wantRef:   "main",
			wantPath:  "file.yaml",
			wantErr:   false,
		},
		{
			name:      "WithWhitespace",
			url:       "  https://github.com/org/repo/blob/main/file.yaml  ",
			wantOwner: "org",
			wantRepo:  "repo",
			wantRef:   "main",
			wantPath:  "file.yaml",
			wantErr:   false,
		},
		{
			name:    "InvalidBlobURLMissingPath",
			url:     "https://github.com/org/repo/blob/main",
			wantErr: true,
		},
		{
			name:    "InvalidBlobURLNotBlob",
			url:     "https://github.com/org/repo/tree/main/path",
			wantErr: true,
		},
		{
			name:    "InvalidRawURLMissingPath",
			url:     "https://raw.githubusercontent.com/org/repo/main",
			wantErr: true,
		},
		{
			name:    "UnsupportedDomain",
			url:     "https://gitlab.com/org/repo/blob/main/file.yaml",
			wantErr: true,
		},
		{
			name:    "EmptyOwner",
			url:     "https://github.com//repo/blob/main/file.yaml",
			wantErr: true,
		},
		{
			name:    "EmptyRepo",
			url:     "https://github.com/owner//blob/main/file.yaml",
			wantErr: true,
		},
		{
			name:    "EmptyPathInBlob",
			url:     "https://github.com/owner/repo/blob/main/",
			wantErr: true,
		},
		{
			name:    "EmptyPathInRaw",
			url:     "https://raw.githubusercontent.com/owner/repo/main/",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, ref, path, err := ParseGitHubURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGitHubURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if owner != tt.wantOwner {
					t.Errorf("ParseGitHubURL() owner = %v, want %v", owner, tt.wantOwner)
				}
				if repo != tt.wantRepo {
					t.Errorf("ParseGitHubURL() repo = %v, want %v", repo, tt.wantRepo)
				}
				if ref != tt.wantRef {
					t.Errorf("ParseGitHubURL() ref = %v, want %v", ref, tt.wantRef)
				}
				if path != tt.wantPath {
					t.Errorf("ParseGitHubURL() path = %v, want %v", path, tt.wantPath)
				}
			}
		})
	}
}

func TestParseCRDMetadata(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    *CRDMetadata
		wantErr bool
	}{
		{
			name: "ValidCRD",
			content: `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: apps.application.giantswarm.io
spec:
  group: application.giantswarm.io
  names:
    kind: App
    plural: apps
    singular: app
  versions:
    - name: v1alpha1
      schema:
        openAPIV3Schema:
          description: App represents an application to deploy.
`,
			want: &CRDMetadata{
				Name:        "apps.application.giantswarm.io",
				Kind:        "App",
				Group:       "application.giantswarm.io",
				Description: "App represents an application to deploy.",
			},
			wantErr: false,
		},
		{
			name: "CRDWithoutDescription",
			content: `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: clusters.cluster.x-k8s.io
spec:
  group: cluster.x-k8s.io
  names:
    kind: Cluster
    plural: clusters
    singular: cluster
  versions:
    - name: v1beta1
`,
			want: &CRDMetadata{
				Name:        "clusters.cluster.x-k8s.io",
				Kind:        "Cluster",
				Group:       "cluster.x-k8s.io",
				Description: "",
			},
			wantErr: false,
		},
		{
			name: "CRDWithMultipleVersions",
			content: `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: machines.cluster.x-k8s.io
spec:
  group: cluster.x-k8s.io
  names:
    kind: Machine
    plural: machines
    singular: machine
  versions:
    - name: v1alpha3
      schema:
        openAPIV3Schema:
          description: Machine is the Schema for the machines API (v1alpha3).
    - name: v1beta1
      schema:
        openAPIV3Schema:
          description: Machine is the Schema for the machines API (v1beta1).
`,
			want: &CRDMetadata{
				Name:        "machines.cluster.x-k8s.io",
				Kind:        "Machine",
				Group:       "cluster.x-k8s.io",
				Description: "Machine is the Schema for the machines API (v1alpha3).", // First version's description
			},
			wantErr: false,
		},
		{
			name: "NotACRD",
			content: `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  key: value
`,
			wantErr: true,
		},
		{
			name: "MissingMetadataName",
			content: `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
spec:
  group: example.com
  names:
    kind: Example
`,
			wantErr: true,
		},
		{
			name: "MissingSpecNamesKind",
			content: `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: examples.example.com
spec:
  group: example.com
  names:
    plural: examples
`,
			wantErr: true,
		},
		{
			name:    "InvalidYAML",
			content: `not: [valid: yaml`,
			wantErr: true,
		},
		{
			name:    "EmptyContent",
			content: "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCRDMetadata(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCRDMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseCRDMetadata() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "WithoutToken",
			config:  Config{},
			wantErr: false,
		},
		{
			name:    "WithToken",
			config:  Config{AuthToken: "ghp_testtoken123"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("New() returned nil service, expected non-nil")
			}
			if !tt.wantErr && got.client == nil {
				t.Errorf("New() service.client is nil, expected non-nil")
			}
		})
	}
}
