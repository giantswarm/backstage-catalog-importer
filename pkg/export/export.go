// Package export takes data from various sources and writes it to a YAML file.
package export

import (
	"bytes"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/giantswarm/backstage-catalog-importer/pkg/catalog"
)

const separator = "---\n"

type Service struct {
	targetPath string
	buffer     bytes.Buffer
}

type Config struct {
	// Path to folder we'll export all files to.
	TargetPath string
}

func New(config Config) *Service {
	s := &Service{
		targetPath: config.TargetPath,
	}

	_, _ = s.buffer.WriteString("#\n# This file was generated automatically. PLEASE DO NOT MODIFY IT BY HAND!\n#\n\n")

	return s
}

// Adds an entity to the export buffer
func (s *Service) AddEntity(entity *catalog.Entity) error {
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
	file, err := os.Create(s.targetPath)
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
