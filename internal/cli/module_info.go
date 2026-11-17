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

const (
	red   = "\x1b[38;5;1m"
	green = "\x1b[38;5;2m"
	reset = "\x1b[0m"
)

func showModuleInfo(inv Invocation, dirs startup.Directories, paths pathresolve.Resolver, repositories []config.Repository, stdout io.Writer) int {
	module, err := operation.ResolveModule(dirs, paths, repositories, inv.Repository, inv.Module, platform.Current())
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}
	configured, err := config.LoadModule(module.SourceDir)
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}
	if configured == nil {
		fmt.Fprintf(stdout, "No configuration is found for module '%s' in repository '%s'.\n", inv.Module, inv.Repository)
		return 0
	}
	fmt.Fprintf(stdout, "Information about module '%s' in repository '%s':\n", inv.Module, inv.Repository)
	if !module.Settings.Available {
		fmt.Fprintf(stdout, "Unavailable on platform '%s'.\n", platform.Current())
		return 0
	}
	fmt.Fprintf(stdout, "Target directory is: %s\n", module.TargetDir)
	if module.Settings.Install != nil {
		fmt.Fprintln(stdout, "Can be installed.")
	}
	if module.Settings.Configure != nil {
		fmt.Fprintln(stdout, "Can be configured.")
	}
	if len(module.Settings.Links) == 0 {
		fmt.Fprintln(stdout, "Has no links.")
		return 0
	}
	fmt.Fprintf(stdout, "Has %d link(s):\n", len(module.Settings.Links))
	for _, mapping := range module.Settings.Links {
		source := filepath.Join(module.SourceDir, filepath.FromSlash(mapping.Source))
		target, err := paths.Resolve(mapping.Target, module.TargetDir, true)
		if err != nil {
			fmt.Fprintln(stdout, err)
			return 1
		}
		fmt.Fprintf(stdout, "%s%s%s -> %s%s%s\n", existenceColor(source), mapping.Source, reset, existenceColor(target), target, reset)
	}
	return 0
}

// Colors show mere filesystem existence, never Alma link ownership.
func existenceColor(path string) string {
	if _, err := os.Lstat(path); err == nil {
		return green
	}
	return red
}
