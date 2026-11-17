// Package config loads Alma's YAML-only configuration.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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
	var node yaml.Node
	if err := yaml.NewDecoder(bytes.NewReader(data)).Decode(&node); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if err := rejectNull(&node); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if len(node.Content) != 1 || node.Content[0].Kind != yaml.MappingNode || !hasKey(node.Content[0], "repositories") {
		return nil, fmt.Errorf("parse %s: expected repositories mapping", path)
	}
	if err := validateRepositoryTypes(node.Content[0]); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	var file repositoryFile
	if err := decoder.Decode(&file); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("multiple YAML documents")
		}
		return nil, fmt.Errorf("parse %s: %w", path, err)
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

func rejectNull(node *yaml.Node) error {
	if node.Tag == "!!null" {
		return fmt.Errorf("explicit YAML null at line %d", node.Line)
	}
	for _, child := range node.Content {
		if err := rejectNull(child); err != nil {
			return err
		}
	}
	return nil
}

// yaml.v3 accepts numeric/bool scalars into Go strings, so validate the YAML
// tags before decoding structs as well as using KnownFields for unknown keys.
func validateRepositoryTypes(root *yaml.Node) error {
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

func hasKey(mapping *yaml.Node, key string) bool {
	for i := 0; i < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return true
		}
	}
	return false
}
