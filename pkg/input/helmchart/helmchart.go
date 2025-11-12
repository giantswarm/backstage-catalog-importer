package helmchart

import (
	"regexp"

	"helm.sh/helm/v4/pkg/chart"
	"sigs.k8s.io/yaml"
)

type Chart struct {
	chart.Metadata
}

func LoadString(content string) (*Chart, error) {
	// Pre-processing: replace [[ *** ]].
	re := regexp.MustCompile(`\[\[[^\]]+]]`)
	content = re.ReplaceAllString(content, "")

	metadata := new(chart.Metadata)
	err := yaml.Unmarshal([]byte(content), metadata)
	if err != nil {
		return nil, err
	}

	return &Chart{Metadata: *metadata}, nil
}
