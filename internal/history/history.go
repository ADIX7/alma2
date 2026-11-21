// Package history persists Go-only link records separately from legacy files.
package history

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ADIX7/alma2/internal/linkplan"
	"gopkg.in/yaml.v3"
)

// Record is the versioned per-source/destination link history.
type Record struct {
	Version      int      `yaml:"version"`
	Source       string   `yaml:"source"`
	Target       string   `yaml:"target"`
	Destinations []string `yaml:"destinations"`
}

// File locates a Go history record without touching old C# metadata filenames.
func File(appDataRoot, source, target string) (string, error) {
	source, err := filepath.Abs(source)
	if err != nil {
		return "", err
	}
	target, err = absoluteTarget(target)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(source + "\x00" + target))
	return filepath.Join(appDataRoot, "history-v1", hex.EncodeToString(sum[:])+".yml"), nil
}

// An empty target identifies a module with no default destination, not the cwd.
func absoluteTarget(target string) (string, error) {
	if target == "" {
		return "", nil
	}
	return filepath.Abs(target)
}

// Save retains prior destinations, deduplicates newly created links, and
// writes a record even when this invocation created zero links.
func Save(appDataRoot, source, target string, created []linkplan.Candidate) error {
	recordFile, err := File(appDataRoot, source, target)
	if err != nil {
		return err
	}
	source, err = filepath.Abs(source)
	if err != nil {
		return err
	}
	target, err = absoluteTarget(target)
	if err != nil {
		return err
	}
	record := Record{Version: 1, Source: source, Target: target, Destinations: []string{}}
	data, err := os.ReadFile(recordFile)
	if err == nil {
		decoder := yaml.NewDecoder(bytes.NewReader(data))
		decoder.KnownFields(true)
		if err := decoder.Decode(&record); err != nil {
			return fmt.Errorf("read history %q: %w", recordFile, err)
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			return fmt.Errorf("history %q contains extra or invalid YAML document: %v", recordFile, err)
		}
		if record.Version != 1 || record.Source != source || record.Target != target {
			return fmt.Errorf("history identity mismatch: %s", recordFile)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read history %q: %w", recordFile, err)
	}
	seen := make(map[string]bool, len(record.Destinations)+len(created))
	for _, destination := range record.Destinations {
		seen[destination] = true
	}
	for _, candidate := range created {
		destination, err := filepath.Abs(candidate.Target)
		if err != nil {
			return err
		}
		if !seen[destination] {
			seen[destination] = true
			record.Destinations = append(record.Destinations, destination)
		}
	}
	data, err = yaml.Marshal(record)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(recordFile), 0700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(recordFile), ".history-*")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Chmod(0600); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporary.Name(), recordFile); err != nil {
		return err
	}
	return nil
}
