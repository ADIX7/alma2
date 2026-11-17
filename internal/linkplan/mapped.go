package linkplan

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/ADIX7/alma2/internal/config"
	"github.com/ADIX7/alma2/internal/pathresolve"
)

// Mapped applies explicit mappings to discovered files and directories.
// A mapped directory becomes one link and suppresses its descendants. When
// multiple mappings name the same source, the last target wins. No filesystem
// mutation occurs. S15 filters payloads before ValidateDestinations is called.
func Mapped(moduleRoot, targetRoot string, links []config.Link, paths pathresolve.Resolver) ([]Candidate, error) {
	mappings := make(map[string]string, len(links))
	for _, link := range links {
		if err := config.ValidateSourcePath(link.Source); err != nil {
			return nil, fmt.Errorf("mapping source: %w", err)
		}
		if link.Target == "" {
			return nil, fmt.Errorf("mapping %q target required", link.Source)
		}
		// Canonicalize aliases such as ./file before applying last-wins.
		mappings[path.Clean(link.Source)] = link.Target
	}
	return plan(moduleRoot, targetRoot, mappings, paths)
}

// ValidateDestinations rejects different effective sources sharing a target.
// It must also run on the final candidate set after S15 payload exclusions,
// before S16/S17 perform any write (including app-data preparation).
func ValidateDestinations(candidates []Candidate) error {
	byDestination := make(map[string]string, len(candidates))
	for _, candidate := range candidates {
		target, err := filepath.Abs(candidate.Target)
		if err != nil {
			return fmt.Errorf("candidate destination %q: %w", candidate.Target, err)
		}
		if source, exists := byDestination[target]; exists && source != candidate.Source {
			return fmt.Errorf("duplicate destination %q for sources %q and %q", target, source, candidate.Source)
		}
		byDestination[target] = candidate.Source
	}
	return nil
}
