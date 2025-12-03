package ociregistry

import (
	"context"
	"testing"

	"github.com/giantswarm/microerror"
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
