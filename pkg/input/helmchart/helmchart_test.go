package helmchart

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"helm.sh/helm/v4/pkg/chart"
)

func TestLoadString(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		name    string
		args    args
		want    *Chart
		wantErr bool
	}{
		{
			name: "Metadata with placeholders in square brackets",
			args: args{
				content: `apiVersion: v2
description: Grafana dashboards accessible by Giant Swarm customers
engine: gotpl
home: https://github.com/giantswarm/dashboards
icon: https://s.giantswarm.io/app-icons/grafana/1/light.svg
name: dashboards
appVersion: [[ .AppVersion ]]
version: [[ .Version ]]
annotations:
  application.giantswarm.io/team: "atlas"
  config.giantswarm.io/version: 1.x.x
dependencies:
  - name: public_dashboards
    version: 1.0.0
`,
			},
			want: &Chart{
				Metadata: chart.Metadata{
					Name:        "dashboards",
					Home:        "https://github.com/giantswarm/dashboards",
					Version:     "",
					Description: "Grafana dashboards accessible by Giant Swarm customers",
					Icon:        "https://s.giantswarm.io/app-icons/grafana/1/light.svg",
					APIVersion:  "v2",
					AppVersion:  "",
					Annotations: map[string]string{
						"application.giantswarm.io/team": "atlas",
						"config.giantswarm.io/version":   "1.x.x",
					},
					Dependencies: []*chart.Dependency{
						{
							Name:    "public_dashboards",
							Version: "1.0.0",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid YAML",
			args: args{
				content: "Just some text",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadString(tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("LoadString() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
