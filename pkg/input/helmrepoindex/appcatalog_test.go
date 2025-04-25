// Package helmrepoindex provides a means to read information form a
// Giant Swarm app catalog, which is technically a Helm repository index
// with conventional metadata.
package helmrepoindex

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		numEntries int
		wantErr    bool
	}{
		{
			name:       "giantswarm catalog",
			path:       "testdata/giantswarm.yaml",
			numEntries: 3,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes, err := os.ReadFile(tt.path)
			if err != nil {
				t.Errorf("Read() error = %v", err)
				return
			}
			f, err := Load(bytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(f.Entries) != tt.numEntries {
				t.Errorf("Read() got %d repositories, want %d", len(f.Entries), tt.numEntries)
			}
		})
	}
}
