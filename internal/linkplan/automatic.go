// Package linkplan discovers candidate symbolic links without writing them.
package linkplan

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ADIX7/alma2/internal/pathresolve"
)

// Candidate is a source/destination pair. RelativeSource uses / separators
// for matching module configuration regardless of the host filesystem.
type Candidate struct {
	Source, Target, RelativeSource string
	Directory                      bool
}

// Automatic plans unmapped files under a nonempty target root, preserving
// relative directory structure. Directories are traversed, never linked by
// default; empty directories yield
// no candidates. Payload exclusions follow in S15.
func Automatic(moduleRoot, targetRoot string) ([]Candidate, error) {
	return plan(moduleRoot, targetRoot, nil, pathresolve.Resolver{})
}

func plan(moduleRoot, targetRoot string, mappings map[string]string, paths pathresolve.Resolver) ([]Candidate, error) {
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
			modulePath := filepath.ToSlash(relativePath)
			mappedTarget, mapped := mappings[modulePath]
			if info.IsDir() && !mapped {
				if err := walk(source, relativePath); err != nil {
					return err
				}
				continue
			}
			if !info.IsDir() && !info.Mode().IsRegular() {
				continue
			}
			var target string
			if mapped {
				target, err = paths.ResolveDestination(mappedTarget, targetRoot)
				if err != nil {
					return fmt.Errorf("mapping %q target: %w", modulePath, err)
				}
			} else if targetRoot != "" {
				target = filepath.Join(targetRoot, relativePath)
			} else {
				continue
			}
			candidates = append(candidates, Candidate{
				Source: source, Target: target, RelativeSource: modulePath,
				Directory: info.IsDir(),
			})
		}
		return nil
	}
	if err := walk(root, ""); err != nil {
		return nil, err
	}
	return candidates, nil
}
