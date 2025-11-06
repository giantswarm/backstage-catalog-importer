// Package export takes bscatalog entities and writes them to YAML files.
package export

import (
	"bytes"
	"cmp"
	"os"
	"slices"

	yaml "go.yaml.in/yaml/v3"

	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/output/bscatalog/v1alpha1"
)

const separator = "---\n"

type Service struct {
	TargetPath string

	collection []*bscatalog.Entity
	buffer     bytes.Buffer
}

type Config struct {
	// Path of the file that will be written.
	TargetPath string
}

// New returns a new export service.
// It is used to collect entities, sort them, and write them to a file.
// For testing convenience, the collection content is stored in a buffer.
// The entity sort order is: APIVersion, Kind, Metadata.Namespace, Metadata.Name.
func New(config Config) *Service {
	s := &Service{
		TargetPath: config.TargetPath,
	}

	return s
}

func (s *Service) updateBuffer() error {
	s.buffer.Reset()
	_, _ = s.buffer.WriteString("#\n# This file was generated automatically. PLEASE DO NOT MODIFY IT BY HAND!\n#\n\n")

	for i := range s.collection {
		yamlBytes, err := yaml.Marshal(&s.collection[i])
		if err != nil {
			return err
		}
		_, err = s.buffer.WriteString(separator)
		if err != nil {
			return err
		}
		_, err = s.buffer.Write(yamlBytes)
		if err != nil {
			return err
		}
	}

	return nil
}

// Adds an entity to the export buffer
func (s *Service) AddEntity(entity *bscatalog.Entity) error {
	s.collection = append(s.collection, entity)

	slices.SortFunc(s.collection, func(a, b *bscatalog.Entity) int {
		return cmp.Or(
			cmp.Compare(a.APIVersion, b.APIVersion),
			cmp.Compare(a.Kind, b.Kind),
			cmp.Compare(a.Metadata.Namespace, b.Metadata.Namespace),
			cmp.Compare(a.Metadata.Name, b.Metadata.Name),
		)
	})

	return nil
}

// Returns the current length of the buffer.
func (s *Service) Len() int {
	_ = s.updateBuffer()
	return s.buffer.Len()
}

// Writes the buffer content to a file.
func (s *Service) WriteFile() error {
	file, err := os.Create(s.TargetPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	err = s.updateBuffer()
	if err != nil {
		return err
	}

	_, err = file.WriteString(s.buffer.String())
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) String() string {
	_ = s.updateBuffer()
	return s.buffer.String()
}
