package installations

import (
	"testing"
)

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
