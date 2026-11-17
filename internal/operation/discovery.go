package operation

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ADIX7/alma2/internal/config"
)

// DiscoverModules lists direct child directories containing .alma.yml, even
// when their platform is unavailable. The repository root and deeper folders
// are not modules; JSON-only directories and symlinked directories are skipped.
func DiscoverModules(root string) ([]string, error) {
	if err := realDirectory(root); err != nil {
		return nil, fmt.Errorf("repository source %q: %w", root, err)
	}
	children, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read repository directory %q: %w", root, err)
	}
	var modules []string
	for _, child := range children {
		if !child.IsDir() { // includes symlinked directories
			continue
		}
		directory := filepath.Join(root, child.Name())
		configuration, err := os.Lstat(filepath.Join(directory, ".alma.yml"))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if configuration.IsDir() {
			return nil, fmt.Errorf("module config %q is a directory", filepath.Join(directory, ".alma.yml"))
		}
		module, err := config.LoadModule(directory)
		if err != nil {
			return nil, err
		}
		if module != nil {
			modules = append(modules, child.Name())
		}
	}
	return modules, nil
}
