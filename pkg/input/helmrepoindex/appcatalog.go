// Package helmrepoindex provides a means to read information form a
// Giant Swarm app catalog, which is technically a Helm repository index
// with conventional metadata.
package helmrepoindex

import (
	"io"
	"net/http"

	"gopkg.in/yaml.v3"
)

// Load returns the repo.IndexFile representing the input YAML.
func Load(b []byte) (*Index, error) {
	i := &Index{}
	err := yaml.Unmarshal(b, i)
	return i, err
}

// LoadFromURL fetches the content from the given URL and returns the
// Index representing the fetched YAML.
func LoadFromURL(url string) (*Index, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return Load(content)
}
