package crdconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "WithReader",
			config: Config{
				Reader: strings.NewReader(""),
			},
			wantErr: false,
		},
		{
			name: "WithFilePath",
			config: Config{
				FilePath: "/some/path.yaml",
			},
			wantErr: false,
		},
		{
			name: "WithBoth",
			config: Config{
				Reader:   strings.NewReader(""),
				FilePath: "/some/path.yaml",
			},
			wantErr: false,
		},
		{
			name:    "WithNeither",
			config:  Config{},
			wantErr: true,
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
		})
	}
}

func TestService_Load(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []Item
		wantErr bool
	}{
		{
			name: "SingleItem",
			input: `- url: https://github.com/org/repo/blob/main/crd.yaml
  owner: team-platform
`,
			want: []Item{
				{
					URL:       "https://github.com/org/repo/blob/main/crd.yaml",
					Owner:     "team-platform",
					Lifecycle: "production", // default
				},
			},
			wantErr: false,
		},
		{
			name: "MultipleItems",
			input: `- url: https://github.com/org/repo/blob/main/crd1.yaml
  owner: team-platform
  lifecycle: experimental
  system: my-system
- url: https://github.com/org/repo/blob/main/crd2.yaml
  owner: team-backend
`,
			want: []Item{
				{
					URL:       "https://github.com/org/repo/blob/main/crd1.yaml",
					Owner:     "team-platform",
					Lifecycle: "experimental",
					System:    "my-system",
				},
				{
					URL:       "https://github.com/org/repo/blob/main/crd2.yaml",
					Owner:     "team-backend",
					Lifecycle: "production", // default
				},
			},
			wantErr: false,
		},
		{
			name: "WithRawGitHubURL",
			input: `- url: https://raw.githubusercontent.com/org/repo/main/crd.yaml
  owner: team-platform
`,
			want: []Item{
				{
					URL:       "https://raw.githubusercontent.com/org/repo/main/crd.yaml",
					Owner:     "team-platform",
					Lifecycle: "production",
				},
			},
			wantErr: false,
		},
		{
			name:    "EmptyInput",
			input:   "",
			want:    nil,
			wantErr: false,
		},
		{
			name: "MissingURL",
			input: `- owner: team-platform
  lifecycle: production
`,
			wantErr: true,
		},
		{
			name: "MissingOwner",
			input: `- url: https://github.com/org/repo/blob/main/crd.yaml
  lifecycle: production
`,
			wantErr: true,
		},
		{
			name: "InvalidYAML",
			input: `- url: [invalid
  owner: team
`,
			wantErr: true,
		},
		{
			name: "AllFieldsSpecified",
			input: `- url: https://github.com/org/repo/blob/main/crd.yaml
  owner: team-platform
  lifecycle: deprecated
  system: legacy-system
`,
			want: []Item{
				{
					URL:       "https://github.com/org/repo/blob/main/crd.yaml",
					Owner:     "team-platform",
					Lifecycle: "deprecated",
					System:    "legacy-system",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := New(Config{Reader: strings.NewReader(tt.input)})
			if err != nil {
				t.Fatalf("New() unexpected error: %v", err)
			}

			got, err := svc.Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Load() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestService_LoadFromFile(t *testing.T) {
	// Create a temporary file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.yaml")

	content := `- url: https://github.com/org/repo/blob/main/crd.yaml
  owner: team-platform
  lifecycle: experimental
`
	if err := os.WriteFile(tmpFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	tests := []struct {
		name     string
		filePath string
		want     []Item
		wantErr  bool
	}{
		{
			name:     "ValidFile",
			filePath: tmpFile,
			want: []Item{
				{
					URL:       "https://github.com/org/repo/blob/main/crd.yaml",
					Owner:     "team-platform",
					Lifecycle: "experimental",
				},
			},
			wantErr: false,
		},
		{
			name:     "NonExistentFile",
			filePath: filepath.Join(tmpDir, "nonexistent.yaml"),
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := New(Config{FilePath: tt.filePath})
			if err != nil {
				t.Fatalf("New() unexpected error: %v", err)
			}

			got, err := svc.Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Load() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestValidateItem(t *testing.T) {
	tests := []struct {
		name    string
		item    Item
		wantErr bool
	}{
		{
			name: "Valid",
			item: Item{
				URL:   "https://github.com/org/repo/blob/main/crd.yaml",
				Owner: "team-platform",
			},
			wantErr: false,
		},
		{
			name: "MissingURL",
			item: Item{
				Owner: "team-platform",
			},
			wantErr: true,
		},
		{
			name: "MissingOwner",
			item: Item{
				URL: "https://github.com/org/repo/blob/main/crd.yaml",
			},
			wantErr: true,
		},
		{
			name:    "MissingBoth",
			item:    Item{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateItem(&tt.item)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name string
		item Item
		want Item
	}{
		{
			name: "EmptyLifecycle",
			item: Item{
				URL:   "https://example.com",
				Owner: "team",
			},
			want: Item{
				URL:       "https://example.com",
				Owner:     "team",
				Lifecycle: "production",
			},
		},
		{
			name: "ExistingLifecycle",
			item: Item{
				URL:       "https://example.com",
				Owner:     "team",
				Lifecycle: "experimental",
			},
			want: Item{
				URL:       "https://example.com",
				Owner:     "team",
				Lifecycle: "experimental",
			},
		},
		{
			name: "WithSystem",
			item: Item{
				URL:    "https://example.com",
				Owner:  "team",
				System: "my-system",
			},
			want: Item{
				URL:       "https://example.com",
				Owner:     "team",
				Lifecycle: "production",
				System:    "my-system",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyDefaults(&tt.item)
			if diff := cmp.Diff(tt.want, tt.item); diff != "" {
				t.Errorf("applyDefaults() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
