package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Module contains ordered platform sections. A nil Module means no .alma.yml.
type Module struct {
	Platforms []Platform `yaml:"platforms"`
}

// Pointer fields retain omission versus an explicitly supplied value for S10.
type Platform struct {
	Selector      string   `yaml:"platform"`
	Target        *string  `yaml:"target"`
	Links         []Link   `yaml:"links"`
	Exclude       []string `yaml:"exclude"`
	ExcludeReadme *bool    `yaml:"excludeReadme"`
	Install       *string  `yaml:"install"`
	Configure     *string  `yaml:"configure"`
}

type Link struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
}

// LoadModule reads only .alma.yml; parsing never creates links or executes commands.
func LoadModule(moduleDir string) (*Module, error) {
	file := filepath.Join(moduleDir, ".alma.yml")
	data, err := os.ReadFile(file)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", file, err)
	}
	var module Module
	if err := decodeStrict(data, file, &module, validateModuleTypes); err != nil {
		return nil, err
	}
	for i, section := range module.Platforms {
		if strings.TrimSpace(section.Selector) == "" {
			return nil, fmt.Errorf("platform %d: selector required", i+1)
		}
		if section.Target != nil {
			if err := yamlPath(*section.Target); err != nil {
				return nil, fmt.Errorf("platform %d target: %w", i+1, err)
			}
		}
		for j, link := range section.Links {
			if err := ValidateSourcePath(link.Source); err != nil {
				return nil, fmt.Errorf("platform %d link %d source: %w", i+1, j+1, err)
			}
			if link.Target == "" {
				return nil, fmt.Errorf("platform %d link %d: target required", i+1, j+1)
			}
			if err := yamlPath(link.Target); err != nil {
				return nil, fmt.Errorf("platform %d link %d target: %w", i+1, j+1, err)
			}
		}
		for _, excluded := range section.Exclude {
			if err := ValidateSourcePath(excluded); err != nil {
				return nil, fmt.Errorf("platform %d exclude: %w", i+1, err)
			}
		}
	}
	return &module, nil
}

func yamlPath(value string) error {
	if strings.Contains(value, `\`) {
		return fmt.Errorf("YAML path must use / separators")
	}
	return nil
}

// ValidateSourcePath checks module-relative YAML source and exclusion paths.
func ValidateSourcePath(value string) error {
	if err := yamlPath(value); err != nil {
		return err
	}
	if value == "" || strings.HasPrefix(value, "/") || (len(value) >= 2 && value[1] == ':') || path.Clean(value) == ".." || strings.HasPrefix(path.Clean(value), "../") {
		return fmt.Errorf("source must be module-relative and cannot escape: %q", value)
	}
	return nil
}

// yaml.v3 can coerce numeric scalars into strings. Inspect YAML node tags and
// shapes before decoding, while KnownFields handles unknown mapping keys.
func validateModuleTypes(root *yaml.Node) error {
	if !hasKey(root, "platforms") {
		return fmt.Errorf("expected platforms mapping")
	}
	for i := 0; i < len(root.Content); i += 2 {
		key, value := root.Content[i], root.Content[i+1]
		if key.Tag != "!!str" {
			return fmt.Errorf("mapping keys must be strings")
		}
		if key.Value != "platforms" {
			continue
		}
		if value.Kind != yaml.SequenceNode {
			return fmt.Errorf("platforms must be a sequence")
		}
		for _, section := range value.Content {
			if section.Kind != yaml.MappingNode {
				return fmt.Errorf("platform section must be a mapping")
			}
			if err := validatePlatformTypes(section); err != nil {
				return err
			}
		}
	}
	return nil
}

func validatePlatformTypes(section *yaml.Node) error {
	for i := 0; i < len(section.Content); i += 2 {
		key, value := section.Content[i], section.Content[i+1]
		if key.Tag != "!!str" {
			return fmt.Errorf("platform field names must be strings")
		}
		switch key.Value {
		case "platform", "target", "install", "configure":
			if value.Kind != yaml.ScalarNode || value.Tag != "!!str" {
				return fmt.Errorf("%s must be a string", key.Value)
			}
		case "excludeReadme":
			if value.Kind != yaml.ScalarNode || value.Tag != "!!bool" {
				return fmt.Errorf("excludeReadme must be a boolean")
			}
		case "exclude", "links":
			if value.Kind != yaml.SequenceNode {
				return fmt.Errorf("%s must be a sequence", key.Value)
			}
			for _, item := range value.Content {
				if key.Value == "exclude" {
					if item.Kind != yaml.ScalarNode || item.Tag != "!!str" {
						return fmt.Errorf("exclude entries must be strings")
					}
				} else if err := validateLinkTypes(item); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validateLinkTypes(link *yaml.Node) error {
	if link.Kind != yaml.MappingNode {
		return fmt.Errorf("link must be a mapping")
	}
	for i := 0; i < len(link.Content); i += 2 {
		key, value := link.Content[i], link.Content[i+1]
		if key.Tag != "!!str" || value.Kind != yaml.ScalarNode || value.Tag != "!!str" {
			return fmt.Errorf("link fields must be strings")
		}
	}
	return nil
}
