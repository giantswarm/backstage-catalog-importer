package api

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/output/bscatalog/v1alpha1"
)

// TestAPI_ToEntity tests the ToEntity method of the API struct.
// This gives us the opportunity to use all setters and options.
func TestAPI_ToEntity(t *testing.T) {
	tests := []struct {
		name    string
		apiName string
		options []Option
		want    *bscatalog.Entity
		wantErr bool
	}{
		{
			name:    "Minimal",
			apiName: "apps.application.giantswarm.io",
			options: []Option{},
			want: &bscatalog.Entity{
				APIVersion: bscatalog.APIVersion,
				Kind:       bscatalog.EntityKindAPI,
				Metadata: bscatalog.EntityMetadata{
					Name: "apps.application.giantswarm.io",
				},
				Spec: bscatalog.APISpec{
					Type:      "crd",
					Lifecycle: "production",
					Owner:     "unspecified",
				},
			},
			wantErr: false,
		},
		{
			name:    "FullFledged",
			apiName: "clusters.cluster.x-k8s.io",
			options: []Option{
				WithNamespace("my-namespace"),
				WithTitle("Cluster"),
				WithDescription("A Kubernetes cluster resource"),
				WithOwner("team-platform"),
				WithType("crd"),
				WithLifecycle("experimental"),
				WithSystem("capi-system"),
				WithDefinition("apiVersion: apiextensions.k8s.io/v1\nkind: CustomResourceDefinition\n..."),
				WithTags("capi", "kubernetes"),
				WithLabels(map[string]string{"category": "infrastructure"}),
			},
			want: &bscatalog.Entity{
				APIVersion: bscatalog.APIVersion,
				Kind:       bscatalog.EntityKindAPI,
				Metadata: bscatalog.EntityMetadata{
					Name:        "clusters.cluster.x-k8s.io",
					Namespace:   "my-namespace",
					Title:       "Cluster",
					Description: "A Kubernetes cluster resource",
					Tags:        []string{"capi", "kubernetes"},
					Labels:      map[string]string{"category": "infrastructure"},
				},
				Spec: bscatalog.APISpec{
					Type:       "crd",
					Lifecycle:  "experimental",
					Owner:      "team-platform",
					System:     "capi-system",
					Definition: "apiVersion: apiextensions.k8s.io/v1\nkind: CustomResourceDefinition\n...",
				},
			},
			wantErr: false,
		},
		{
			name:    "WithCustomAnnotationsAndLinks",
			apiName: "custom-api",
			options: []Option{},
			want: &bscatalog.Entity{
				APIVersion: bscatalog.APIVersion,
				Kind:       bscatalog.EntityKindAPI,
				Metadata: bscatalog.EntityMetadata{
					Name: "custom-api",
					Annotations: map[string]string{
						"custom.io/annotation": "custom-value",
					},
					Links: []bscatalog.EntityLink{
						{
							URL:   "https://docs.example.com/api",
							Title: "API Documentation",
							Icon:  "docs",
							Type:  "documentation",
						},
					},
				},
				Spec: bscatalog.APISpec{
					Type:      "crd",
					Lifecycle: "production",
					Owner:     "unspecified",
				},
			},
			wantErr: false,
		},
		{
			name:    "OpenAPIType",
			apiName: "petstore-api",
			options: []Option{
				WithType("openapi"),
				WithOwner("team-backend"),
				WithDescription("The Petstore API"),
			},
			want: &bscatalog.Entity{
				APIVersion: bscatalog.APIVersion,
				Kind:       bscatalog.EntityKindAPI,
				Metadata: bscatalog.EntityMetadata{
					Name:        "petstore-api",
					Description: "The Petstore API",
				},
				Spec: bscatalog.APISpec{
					Type:      "openapi",
					Lifecycle: "production",
					Owner:     "team-backend",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api, err := New(tt.apiName, tt.options...)
			if err != nil && !tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Apply custom modifications for specific test cases
			if tt.name == "WithCustomAnnotationsAndLinks" {
				api.SetAnnotation("custom.io/annotation", "custom-value")
				api.AddLink(bscatalog.EntityLink{
					URL:   "https://docs.example.com/api",
					Title: "API Documentation",
					Icon:  "docs",
					Type:  "documentation",
				})
			}

			got := api.ToEntity()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("API.ToEntity() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		name    string
		options []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *API
		wantErr bool
	}{
		{
			name: "Success",
			args: args{name: "apps.application.giantswarm.io"},
			want: &API{
				Name:      "apps.application.giantswarm.io",
				Namespace: "default",
				Owner:     "unspecified",
				Type:      "crd",
				Lifecycle: "production",
			},
		},
		{
			name:    "EmptyName",
			args:    args{name: ""},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.name, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("New() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestGeneric tests API creation, adders, setters, and options in flexible ways.
func TestGeneric(t *testing.T) {
	tests := []struct {
		name    string
		code    func() (*API, error)
		want    *API
		wantErr bool
	}{
		{
			name: "SimpleSuccess",
			code: func() (*API, error) {
				return New("minimal-api")
			},
			want: &API{
				Name:      "minimal-api",
				Namespace: "default",
				Owner:     "unspecified",
				Type:      "crd",
				Lifecycle: "production",
			},
		},
		{
			name: "SomeSetters",
			code: func() (*API, error) {
				a, _ := New("test-api")
				a.AddTag("tag1")
				a.AddLink(bscatalog.EntityLink{
					Title: "link1",
					Type:  "documentation",
					URL:   "https://example.com",
					Icon:  "docs",
				})
				a.SetAnnotation("key1", "value1")
				a.SetLabel("label1", "value1")
				return a, nil
			},
			want: &API{
				Annotations: map[string]string{"key1": "value1"},
				Labels:      map[string]string{"label1": "value1"},
				Lifecycle:   "production",
				Links: []bscatalog.EntityLink{
					{URL: "https://example.com", Title: "link1", Icon: "docs", Type: "documentation"},
				},
				Name:      "test-api",
				Namespace: "default",
				Owner:     "unspecified",
				Tags:      []string{"tag1"},
				Type:      "crd",
			},
		},
		{
			name: "WithAllOptions",
			code: func() (*API, error) {
				return New("full-api",
					WithNamespace("custom-ns"),
					WithTitle("Full API"),
					WithDescription("A fully configured API"),
					WithOwner("team-api"),
					WithType("grpc"),
					WithLifecycle("deprecated"),
					WithSystem("api-system"),
					WithDefinition("syntax = \"proto3\";"),
					WithTags("grpc", "backend"),
					WithLabels(map[string]string{"team": "api"}),
				)
			},
			want: &API{
				Name:        "full-api",
				Namespace:   "custom-ns",
				Title:       "Full API",
				Description: "A fully configured API",
				Owner:       "team-api",
				Type:        "grpc",
				Lifecycle:   "deprecated",
				System:      "api-system",
				Definition:  "syntax = \"proto3\";",
				Tags:        []string{"grpc", "backend"},
				Labels:      map[string]string{"team": "api"},
			},
		},
		{
			name: "EmptyOptionsIgnored",
			code: func() (*API, error) {
				return New("test-api",
					WithNamespace(""), // Should not change default
					WithOwner(""),     // Should not change default
					WithType(""),      // Should not change default
					WithLifecycle(""), // Should not change default
				)
			},
			want: &API{
				Name:      "test-api",
				Namespace: "default",
				Owner:     "unspecified",
				Type:      "crd",
				Lifecycle: "production",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.code()
			if (err != nil) != tt.wantErr {
				t.Errorf("code() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("code() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
