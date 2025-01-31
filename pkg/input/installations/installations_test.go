// Reads Giant Swarm installations info from the giantswarm/installations repository
package installations

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_parseInstallationInfo(t *testing.T) {

	tests := []struct {
		name      string
		inputPath string
		want      *Installation
		wantErr   bool
	}{
		{
			name:      "akita",
			inputPath: "testdata/akita.yaml",
			want: &Installation{
				AccountEngineer:        "Jane Doe",
				AccountEngineersHandle: "ae-acme",
				Aws: &AwsDetails{
					Region: "cn-northwest-1",
					HostCluster: AwsIdentity{
						Account: "123456789012",
					},
					GuestCluster: AwsIdentity{
						Account: "123456789012",
					},
				},
				Base:             "akita.inc.example.com",
				CcrRepository:    "acme-configs",
				CmcRepository:    "acme-management-clusters",
				Codename:         "akita",
				Customer:         "acme",
				EscalationMatrix: "Here is some information regarding how to escalate incidents\n",
				Pipeline:         "stable",
				Provider:         "aws",
				Region:           "cn-northwest-1",
				Slack:            &SlackDetails{Support: []string{"support-acme-admin"}},
			},
			wantErr: false,
		},
		{
			name:      "alba.yaml",
			inputPath: "testdata/alba.yaml",
			want: &Installation{
				Base:                   "capi.aws.k8s.example.com",
				Codename:               "alba",
				Customer:               "acme",
				CmcRepository:          "acme-management-clusters",
				CcrRepository:          "acme-configs",
				AccountEngineer:        "Jane Doe",
				AccountEngineersHandle: "ae-acme",
				EscalationMatrix:       "Here is some information regarding how to escalate incidents\n",
				Slack:                  &SlackDetails{Support: []string{"support-acme-admin"}},
				Pipeline:               "stable",
				Provider:               "capa",
				Region:                 "eu-west-1",
				Aws: &AwsDetails{
					Region: "eu-west-1",
					HostCluster: AwsIdentity{
						Account:      "123456789012",
						AdminRoleARN: "arn:aws:iam::123456789012:role/GiantSwarmAdmin",
					},
					GuestCluster: AwsIdentity{
						Account: "123456789012",
					},
				},
			},
			wantErr: false,
		},
		{
			name:      "goose.yaml",
			inputPath: "testdata/goose.yaml",
			want: &Installation{
				AccountEngineer:        "Team Phoenix",
				AccountEngineersHandle: "support-phoenix",
				Base:                   "azuretest.gigantic.io",
				CcrRepository:          "giantswarm-configs",
				CmcRepository:          "giantswarm-management-clusters",
				Codename:               "goose",
				Customer:               "giantswarm",
				Pipeline:               "ephemeral",
				Provider:               "capz",
				Region:                 "fantasy-region",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read the file
			content, err := os.ReadFile(tt.inputPath)
			if err != nil {
				t.Fatalf("Error reading file %s: %s", tt.inputPath, err)
			}

			got, err := parseInstallationInfo(content)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInstallationInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("parseInstallationInfo() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func FuzzParseInstallationInfo(f *testing.F) {
	// Add valid test cases from existing tests as seed corpus
	f.Add([]byte(`
codename: akita
customer: acme
provider: aws
region: cn-northwest-1
base: akita.inc.example.com
pipeline: stable
accountEngineer: Jane Doe
accountEngineersHandle: ae-acme
ccrRepository: acme-configs
cmcRepository: acme-management-clusters
`))

	// Add minimal valid case
	f.Add([]byte(`codename: test
provider: aws`))

	// Add empty input
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parseInstallationInfo(data)
	})
}
