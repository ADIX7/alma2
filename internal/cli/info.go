package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"

	"github.com/ADIX7/alma2/internal/config"
	"github.com/ADIX7/alma2/internal/operation"
	"github.com/ADIX7/alma2/internal/pathresolve"
	"github.com/ADIX7/alma2/internal/platform"
	"github.com/ADIX7/alma2/internal/startup"
)

func runInfo(inv Invocation, stdout io.Writer) int {
	if inv.Module != "" {
		fmt.Fprintln(stdout, "Command not implemented: module info")
		return 1
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}
	inputs := startup.CurrentInputs(cwd)
	dirs := startup.Resolve(inputs)
	for _, warning := range dirs.Warnings {
		fmt.Fprintln(stdout, warning)
	}
	repositories, err := config.LoadRepositories(dirs.ConfigRoot)
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}
	paths := pathresolve.Current()
	if inv.Repository == "" {
		version := "unknown"
		if build, ok := debug.ReadBuildInfo(); ok && build.Main.Version != "" {
			version = build.Main.Version
		}
		fmt.Fprintf(stdout, "Alma %s\n\nAppData folder: %s\n", version, dirs.AppDataRoot)
		if dirs.ConfigRoot != "" {
			fmt.Fprintf(stdout, "Configuration folder: %s\n", dirs.ConfigRoot)
		} else {
			fmt.Fprintln(stdout, "Configuration folder not exists.")
			if inputs.Home != "" {
				fmt.Fprintf(stdout, "Preferred configuration folder is: %s\n", filepath.Join(inputs.Home, ".config", "alma"))
			}
		}
		fmt.Fprintf(stdout, "\nPlatform is '%s'\n\n", platform.Current())
		if len(repositories) == 0 {
			fmt.Fprintln(stdout, "No repositories found")
			return 0
		}
		fmt.Fprintln(stdout, "Repositories:")
		for _, entry := range repositories {
			repo, err := operation.ResolveRepository(dirs, paths, repositories, entry.Name)
			if err != nil {
				fmt.Fprintln(stdout, err)
				return 1
			}
			if info, err := os.Stat(repo.SourceRoot); err != nil || !info.IsDir() {
				fmt.Fprintf(stdout, "%s (containing folder not exists %s)\n", entry.Name, repo.SourceRoot)
			} else {
				fmt.Fprintln(stdout, entry.Name)
			}
		}
		return 0
	}
	repo, err := operation.ResolveRepository(dirs, paths, repositories, inv.Repository)
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}
	names, err := operation.DiscoverModules(repo.SourceRoot)
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}
	sort.Strings(names)
	fmt.Fprintf(stdout, "Repository '%s' contains %d modules:\n\n", inv.Repository, len(names))
	for _, name := range names {
		module, err := config.LoadModule(filepath.Join(repo.SourceRoot, name))
		if err != nil {
			fmt.Fprintln(stdout, err)
			return 1
		}
		settings := config.Select(module, platform.Current())
		if !settings.Available {
			fmt.Fprintf(stdout, "%s  [Unavailable]\n", name)
			continue
		}
		count := len(settings.Links)
		fmt.Fprintf(stdout, "%s  %d link(s)", name, count)
		if settings.Install != nil {
			fmt.Fprint(stdout, "    [Install]")
		}
		if settings.Configure != nil {
			fmt.Fprint(stdout, "    [Configure]")
		}
		fmt.Fprintln(stdout)
	}
	return 0
}
