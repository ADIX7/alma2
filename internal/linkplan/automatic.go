// Package linkplan discovers candidate symbolic links without writing them.
package linkplan

import (
	"fmt"
	"os"
	"path/filepath"
)

// Candidate is a source/destination pair. RelativeSource uses / separators
// for matching module configuration regardless of the host filesystem.
type Candidate struct {
	Source, Target, RelativeSource string
	Directory                      bool
}

// Automatic plans unmapped files, preserving relative directory structure.
// Directories are traversed, never linked by default; empty directories yield
// no candidates. Payload exclusion and explicit mappings follow in S14–S15.
func Automatic(moduleRoot, targetRoot string) ([]Candidate, error) {
	root, err := filepath.Abs(moduleRoot)
	if err != nil {
		return nil, fmt.Errorf("module root: %w", err)
	}
	info, err := os.Lstat(root)
	if err != nil {
		return nil, fmt.Errorf("module root %q: %w", root, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return nil, fmt.Errorf("module root %q is not a real directory", root)
	}
	var candidates []Candidate
	var walk func(string, string) error
	walk = func(directory, relative string) error {
		children, err := os.ReadDir(directory)
		if err != nil {
			return fmt.Errorf("read module directory %q: %w", directory, err)
		}
		for _, child := range children {
			source := filepath.Join(directory, child.Name())
			info, err := os.Lstat(source)
			if err != nil {
				return fmt.Errorf("inspect module source %q: %w", source, err)
			}
			if info.Mode()&os.ModeSymlink != 0 {
				continue
			}
			relativePath := filepath.Join(relative, child.Name())
			if info.IsDir() {
				if err := walk(source, relativePath); err != nil {
					return err
				}
				continue
			}
			if !info.Mode().IsRegular() {
				continue
			}
			candidates = append(candidates, Candidate{
				Source: source, Target: filepath.Join(targetRoot, relativePath),
				RelativeSource: filepath.ToSlash(relativePath),
			})
		}
		return nil
	}
	if err := walk(root, ""); err != nil {
		return nil, err
	}
	return candidates, nil
}
