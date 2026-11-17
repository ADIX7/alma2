package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ADIX7/alma2/internal/config"
	"github.com/ADIX7/alma2/internal/operation"
	"github.com/ADIX7/alma2/internal/pathresolve"
	"github.com/ADIX7/alma2/internal/platform"
	"github.com/ADIX7/alma2/internal/startup"
)

func runReadme(inv Invocation, stdout io.Writer) int {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}
	dirs := startup.Resolve(startup.CurrentInputs(cwd))
	for _, warning := range dirs.Warnings {
		fmt.Fprintln(stdout, warning)
	}
	repositories, err := config.LoadRepositories(dirs.ConfigRoot)
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}
	module, err := operation.ResolveModule(dirs, pathresolve.Current(), repositories, inv.Repository, inv.Module, platform.Current())
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}
	entries, err := os.ReadDir(module.SourceDir)
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}
	// Compare directory-entry names exactly, even on case-insensitive hosts.
	for _, name := range []string{"README.md", "README.txt", "README"} {
		for _, entry := range entries {
			if entry.Name() != name || entry.IsDir() {
				continue
			}
			contents, err := os.ReadFile(filepath.Join(module.SourceDir, name))
			if err != nil {
				fmt.Fprintln(stdout, err)
				return 1
			}
			if _, err := stdout.Write(contents); err != nil {
				return 1
			}
			return 0
		}
	}
	fmt.Fprintln(stdout, "No README file found. Supported formats: README.md, README.txt, README")
	return 0
}
