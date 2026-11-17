// Package operation resolves repository and module paths without filesystem writes.
package operation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ADIX7/alma2/internal/config"
	"github.com/ADIX7/alma2/internal/pathresolve"
	"github.com/ADIX7/alma2/internal/startup"
)

// Repository holds source and default destination roots. RepositoryPathPresent
// distinguishes an omitted source field for ls, which requires that field.
type Repository struct {
	SourceRoot, TargetRoot string
	RepositoryPathPresent  bool
}

// Module holds the selected module's roots and applicable settings.
type Module struct {
	Repository
	Name, SourceDir, TargetDir string
	Settings                   config.Settings
}

// ResolveRepository uses working-directory fallbacks only when no named
// repository is requested or when its corresponding path field is omitted.
func ResolveRepository(dirs startup.Directories, paths pathresolve.Resolver, repositories []config.Repository, name string) (Repository, error) {
	result := Repository{SourceRoot: dirs.WorkDir, TargetRoot: filepath.Dir(dirs.WorkDir)}
	if name == "" {
		return result, nil
	}
	for _, entry := range repositories {
		if entry.Name != name {
			continue
		}
		if entry.RepositoryPath != nil {
			resolved, err := paths.Resolve(*entry.RepositoryPath, dirs.WorkDir, true)
			if err != nil {
				return Repository{}, fmt.Errorf("repository %q source: %w", name, err)
			}
			result.SourceRoot = resolved
			result.RepositoryPathPresent = true
		}
		if entry.LinkPath != nil {
			resolved, err := paths.Resolve(*entry.LinkPath, dirs.WorkDir, true)
			if err != nil {
				return Repository{}, fmt.Errorf("repository %q destination: %w", name, err)
			}
			result.TargetRoot = resolved
		}
		return result, nil
	}
	return Repository{}, fmt.Errorf("unknown repository: %s", name)
}

// ResolveModule verifies the selected source's existence and boundary, then
// selects its YAML settings. Targets are resolved but never created here.
func ResolveModule(dirs startup.Directories, paths pathresolve.Resolver, repositories []config.Repository, repositoryName, moduleName, identifier string) (Module, error) {
	repo, err := ResolveRepository(dirs, paths, repositories, repositoryName)
	if err != nil {
		return Module{}, err
	}
	if !safeModuleName(moduleName) {
		return Module{}, fmt.Errorf("unsafe module name: %q", moduleName)
	}
	moduleDir, err := moduleDirectory(repo.SourceRoot, moduleName)
	if err != nil {
		return Module{}, err
	}
	configuration, err := config.LoadModule(moduleDir)
	if err != nil {
		return Module{}, err
	}
	settings := config.Select(configuration, identifier)
	target := repo.TargetRoot
	if settings.Target != nil {
		target, err = paths.Resolve(*settings.Target, target, true)
		if err != nil {
			return Module{}, fmt.Errorf("module %q destination: %w", moduleName, err)
		}
	}
	return Module{Repository: repo, Name: moduleName, SourceDir: moduleDir, TargetDir: target, Settings: settings}, nil
}

func safeModuleName(name string) bool {
	if name == "" || name == "." || strings.ContainsAny(name, `/\`) || (len(name) >= 2 && name[1] == ':') {
		return false
	}
	// Only a single direct-child directory name is valid.
	return filepath.IsLocal(filepath.FromSlash(name))
}

func moduleDirectory(root, name string) (string, error) {
	if root == "" {
		return "", fmt.Errorf("source directory is not available")
	}
	current := filepath.Clean(root)
	if err := realDirectory(current); err != nil {
		return "", fmt.Errorf("source directory %q: %w", current, err)
	}
	current = filepath.Join(current, name)
	if err := realDirectory(current); err != nil {
		return "", fmt.Errorf("module directory %q: %w", current, err)
	}
	return current, nil
}

func realDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("symlinked directory is not allowed")
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory")
	}
	return nil
}
