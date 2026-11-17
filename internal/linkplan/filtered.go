package linkplan

import (
	"fmt"
	"path"
	"strings"

	"github.com/ADIX7/alma2/internal/config"
	"github.com/ADIX7/alma2/internal/pathresolve"
)

// Plan applies mappings and source exclusions, then rejects duplicate final
// destinations. It is read-only; callers must finish planning before creating
// app-data, target directories, links, or history.
func Plan(moduleRoot, targetRoot string, settings config.Settings, paths pathresolve.Resolver) ([]Candidate, error) {
	if !settings.Available {
		return nil, fmt.Errorf("module is unavailable on this platform")
	}
	candidates, err := Mapped(moduleRoot, targetRoot, settings.Links, paths)
	if err != nil {
		return nil, err
	}
	exclusions := make([]string, 0, len(settings.Exclude))
	for _, excluded := range settings.Exclude {
		if err := config.ValidateSourcePath(excluded); err != nil {
			return nil, fmt.Errorf("exclude: %w", err)
		}
		exclusions = append(exclusions, path.Clean(excluded))
	}
	filtered := make([]Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		relative := candidate.RelativeSource
		// Configuration files are never eligible, even if explicitly mapped.
		if protectedConfig(path.Base(relative)) || (settings.ExcludeReadme && !candidate.Directory && rootReadme(relative)) {
			continue
		}
		excluded := false
		for _, name := range exclusions {
			if name == "." || relative == name || strings.HasPrefix(relative, name+"/") {
				excluded = true
				break
			}
		}
		if !excluded {
			filtered = append(filtered, candidate)
		}
	}
	if err := ValidateDestinations(filtered); err != nil {
		return nil, err
	}
	return filtered, nil
}

func protectedConfig(name string) bool {
	return name == ".alma.yml" || name == ".alma-config.json"
}

func rootReadme(relative string) bool {
	return relative == "README.md" || relative == "README.txt" || relative == "README"
}
