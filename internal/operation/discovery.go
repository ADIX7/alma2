package operation

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ADIX7/alma2/internal/config"
)

// DiscoverModules returns directories containing .alma.yml, including an
// unavailable configured module or the repository root (reported as ".").
// JSON-only directories are not modules. No source symlink is traversed.
func DiscoverModules(root string) ([]string, error) {
	if err := realDirectory(root); err != nil {
		return nil, fmt.Errorf("repository source %q: %w", root, err)
	}
	var modules []string
	var walk func(string, string) error
	walk = func(directory, name string) error {
		configuration, err := os.Lstat(filepath.Join(directory, ".alma.yml"))
		if err == nil {
			if configuration.IsDir() {
				return fmt.Errorf("module config %q is a directory", filepath.Join(directory, ".alma.yml"))
			}
			module, err := config.LoadModule(directory)
			if err != nil {
				return err
			}
			if module != nil {
				modules = append(modules, name)
			}
		} else if !os.IsNotExist(err) {
			return err
		}
		children, err := os.ReadDir(directory)
		if err != nil {
			return fmt.Errorf("read repository directory %q: %w", directory, err)
		}
		for _, child := range children {
			if !child.IsDir() { // includes symlinked directories
				continue
			}
			next := child.Name()
			if name != "." {
				next = name + "/" + next
			}
			if err := walk(filepath.Join(directory, child.Name()), next); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(root, "."); err != nil {
		return nil, err
	}
	return modules, nil
}
