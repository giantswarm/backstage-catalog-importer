package ociregistry

import (
	"context"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/google/go-cmp/cmp"
)

func TestNewRegistry(t *testing.T) {
	type args struct {
		ctx    context.Context
		config Config
	}
	tests := []struct {
		name        string
		args        args
		want        *Registry
		wantErr     bool
		wantErrType error
	}{
		{
			name: "Valid configuration with hostname",
			args: args{
				ctx: context.Background(),
				config: Config{
					Hostname: "registry.example.com",
				},
			},
			want:    &Registry{}, // We'll check that Client is not nil separately
			wantErr: false,
		},
		{
			name: "Valid configuration with Docker Hub",
			args: args{
				ctx: context.Background(),
				config: Config{
					Hostname: "docker.io",
				},
			},
			want:    &Registry{}, // We'll check that Client is not nil separately
			wantErr: false,
		},
		{
			name: "Valid configuration with localhost",
			args: args{
				ctx: context.Background(),
				config: Config{
					Hostname: "localhost:5000",
				},
			},
			want:    &Registry{}, // We'll check that Client is not nil separately
			wantErr: false,
		},
		{
			name: "Invalid configuration - empty hostname",
			args: args{
				ctx: context.Background(),
				config: Config{
					Hostname: "",
				},
			},
			want:        nil,
			wantErr:     true,
			wantErrType: invalidConfigError,
		},
		{
			name: "Invalid configuration - whitespace only hostname",
			args: args{
				ctx: context.Background(),
				config: Config{
					Hostname: "   ",
				},
			},
			want:        nil,
			wantErr:     true,
			wantErrType: couldNotCreateRegistryClientError,
		},
		{
			name: "Invalid configuration - invalid hostname format",
			args: args{
				ctx: context.Background(),
				config: Config{
					Hostname: "invalid://hostname",
				},
			},
			want:        nil,
			wantErr:     true,
			wantErrType: couldNotCreateRegistryClientError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRegistry(tt.args.ctx, tt.args.config)

			// Check error expectations
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRegistry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check specific error type if expected
			if tt.wantErr && tt.wantErrType != nil {
				if microerror.Cause(err) != tt.wantErrType {
					t.Errorf("NewRegistry() error type = %v, want %v", microerror.Cause(err), tt.wantErrType)
				}
			}

			// For successful cases, verify the registry is properly initialized
			if !tt.wantErr {
				if got == nil {
					t.Errorf("NewRegistry() returned nil registry for valid config")
					return
				}
				if got.Client == nil {
					t.Errorf("NewRegistry() returned registry with nil Client")
				}
			}

			// For error cases, ensure nil is returned
			if tt.wantErr && got != nil {
				t.Errorf("NewRegistry() returned non-nil registry for invalid config")
			}
		})
	}
}

func TestSortTagsBySemver(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Basic semver sorting",
			input:    []string{"1.0.0", "2.0.0", "1.5.0"},
			expected: []string{"2.0.0", "1.5.0", "1.0.0"},
		},
		{
			name:     "Semver with v prefix",
			input:    []string{"v2.0.0", "v1.0.0", "v1.5.0"},
			expected: []string{"v2.0.0", "v1.5.0", "v1.0.0"},
		},
		{
			name:     "Releases sort before pre-releases",
			input:    []string{"1.0.0", "1.0.0-alpha", "1.0.0-beta", "1.0.0-rc1"},
			expected: []string{"1.0.0", "1.0.0-rc1", "1.0.0-beta", "1.0.0-alpha"},
		},
		{
			name:     "Complex pre-release sorting",
			input:    []string{"2.0.0", "1.0.0", "1.0.0-alpha.1", "1.0.0-alpha.2", "1.0.0-beta"},
			expected: []string{"2.0.0", "1.0.0", "1.0.0-beta", "1.0.0-alpha.2", "1.0.0-alpha.1"},
		},
		{
			name:     "Non-semver tags sorted reverse alphabetically after semver",
			input:    []string{"latest", "1.0.0", "main", "0.1.0"},
			expected: []string{"1.0.0", "0.1.0", "main", "latest"},
		},
		{
			name:     "Mixed semver and non-semver with pre-releases",
			input:    []string{"latest", "1.0.0", "1.0.0-alpha", "dev", "0.9.0"},
			expected: []string{"1.0.0", "1.0.0-alpha", "0.9.0", "latest", "dev"},
		},
		{
			name:     "All non-semver tags",
			input:    []string{"latest", "main", "dev", "alpha"},
			expected: []string{"main", "latest", "dev", "alpha"},
		},
		{
			name:     "Empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "Single element",
			input:    []string{"1.0.0"},
			expected: []string{"1.0.0"},
		},
		{
			name:     "Patch version sorting",
			input:    []string{"1.0.10", "1.0.2", "1.0.1"},
			expected: []string{"1.0.10", "1.0.2", "1.0.1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying the original
			input := make([]string, len(tt.input))
			copy(input, tt.input)

			sortTagsBySemver(input)

			if diff := cmp.Diff(tt.expected, input); diff != "" {
				t.Errorf("sortTagsBySemver() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
