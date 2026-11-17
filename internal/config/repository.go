// Package config loads Alma's YAML-only configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Repository preserves declaration order; paths are optional and resolved by
// the operation context, not during configuration loading.
type Repository struct {
	Name           string  `yaml:"name"`
	RepositoryPath *string `yaml:"repositoryPath"`
	LinkPath       *string `yaml:"linkPath"`
}

type repositoryFile struct {
	Repositories []Repository `yaml:"repositories"`
}

// LoadRepositories reads only repository.yml under configRoot. A missing file
// gives an empty list; malformed or explicitly null YAML is an error.
func LoadRepositories(configRoot string) ([]Repository, error) {
	if configRoot == "" {
		return []Repository{}, nil
	}
	path := filepath.Join(configRoot, "repository.yml")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return []Repository{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var file repositoryFile
	if err := decodeStrict(data, path, &file, validateRepositoryTypes); err != nil {
		return nil, err
	}
	for i, repository := range file.Repositories {
		if strings.TrimSpace(repository.Name) == "" {
			return nil, fmt.Errorf("repository %d: name required", i+1)
		}
		for name, value := range map[string]*string{"repositoryPath": repository.RepositoryPath, "linkPath": repository.LinkPath} {
			if value != nil && strings.Contains(*value, `\`) {
				return nil, fmt.Errorf("repository %d: %s must use / separators", i+1, name)
			}
		}
	}
	return file.Repositories, nil
}

// yaml.v3 accepts numeric/bool scalars into Go strings, so validate the YAML
// tags before decoding structs as well as using KnownFields for unknown keys.
func validateRepositoryTypes(root *yaml.Node) error {
	if !hasKey(root, "repositories") {
		return fmt.Errorf("expected repositories mapping")
	}
	for i := 0; i < len(root.Content); i += 2 {
		key, value := root.Content[i], root.Content[i+1]
		if key.Tag != "!!str" {
			return fmt.Errorf("mapping keys must be strings")
		}
		if key.Value != "repositories" {
			continue // KnownFields reports the unknown field.
		}
		if value.Kind != yaml.SequenceNode {
			return fmt.Errorf("repositories must be a sequence")
		}
		for _, entry := range value.Content {
			if entry.Kind != yaml.MappingNode {
				return fmt.Errorf("repository entries must be mappings")
			}
			for j := 0; j < len(entry.Content); j += 2 {
				field, item := entry.Content[j], entry.Content[j+1]
				if field.Tag != "!!str" {
					return fmt.Errorf("repository field names must be strings")
				}
				if item.Kind != yaml.ScalarNode || item.Tag != "!!str" {
					return fmt.Errorf("repository %s must be a string", field.Value)
				}
			}
		}
	}
	return nil
}
