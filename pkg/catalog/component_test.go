package catalog

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestComponent_ToEntity(t *testing.T) {

	tests := []struct {
		name          string
		componentName string
		options       []Option
		want          *Entity
		wantErr       bool
	}{
		{
			name:          "Minimal",
			componentName: "minimal",
			options:       []Option{},
			want: &Entity{
				APIVersion: "backstage.io/v1alpha1",
				Kind:       EntityKindComponent,
				Metadata: EntityMetadata{
					Name: "minimal",
				},
				Spec: ComponentSpec{
					Type:      "unspecified",
					Lifecycle: "production",
					Owner:     "unspecified",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component, err := NewComponent(tt.componentName, tt.options...)
			if err != nil && !tt.wantErr {
				t.Errorf("NewComponent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got := component.ToEntity()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Component.ToEntity() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
