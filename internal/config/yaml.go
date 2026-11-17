package config

import (
	"bytes"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

func decodeStrict(data []byte, path string, out any, validate func(*yaml.Node) error) error {
	var node yaml.Node
	if err := yaml.NewDecoder(bytes.NewReader(data)).Decode(&node); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	if err := rejectNull(&node); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	if len(node.Content) != 1 || node.Content[0].Kind != yaml.MappingNode {
		return fmt.Errorf("parse %s: expected YAML mapping", path)
	}
	if err := validate(node.Content[0]); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("multiple YAML documents")
		}
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
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

func hasKey(mapping *yaml.Node, key string) bool {
	for i := 0; i < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return true
		}
	}
	return false
}
