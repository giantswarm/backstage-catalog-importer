// Package export takes bscatalog entities and writes them to YAML files.
package export

import (
	"bytes"
	"os"

	"gopkg.in/yaml.v3"

	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/bscatalog/v1alpha1"
)

const separator = "---\n"

type Service struct {
	TargetPath string
	buffer     bytes.Buffer
}

type Config struct {
	// Path of the file that will be written.
	TargetPath string
}

func New(config Config) *Service {
	s := &Service{
		TargetPath: config.TargetPath,
	}

	_, _ = s.buffer.WriteString("#\n# This file was generated automatically. PLEASE DO NOT MODIFY IT BY HAND!\n#\n\n")

	return s
}

// Adds an entity to the export buffer
func (s *Service) AddEntity(entity *bscatalog.Entity) error {
	yamlBytes, err := yaml.Marshal(&entity)
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

	return nil
}

// Returns the current length of the buffer.
func (s *Service) Len() int {
	return s.buffer.Len()
}

// Writes the buffer content to a file.
func (s *Service) WriteFile() error {
	file, err := os.Create(s.TargetPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(s.buffer.String())
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) String() string {
	return s.buffer.String()
}
