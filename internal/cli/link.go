package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/ADIX7/alma2/internal/config"
	"github.com/ADIX7/alma2/internal/history"
	"github.com/ADIX7/alma2/internal/linkapply"
	"github.com/ADIX7/alma2/internal/linkplan"
	"github.com/ADIX7/alma2/internal/operation"
	"github.com/ADIX7/alma2/internal/pathresolve"
	"github.com/ADIX7/alma2/internal/platform"
	"github.com/ADIX7/alma2/internal/startup"
)

func runLink(inv Invocation, stdout io.Writer) int {
	if inv.LinkHelp {
		fmt.Fprintln(stdout, "Usage: alma link [repository] <module> [-d|--dry-run]")
		return 0
	}
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
	paths := pathresolve.Current()
	module, err := operation.ResolveModule(dirs, paths, repositories, inv.Repository, inv.Module, platform.Current())
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}
	candidates, err := linkplan.Plan(module.SourceDir, module.TargetDir, module.Settings, paths)
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}
	if inv.Repository == "" && len(candidates) == 0 {
		if inv.DryRun {
			fmt.Fprint(stdout, "Dry run. ")
		}
		fmt.Fprintln(stdout, "No links to create.")
		return 0
	}
	if !inv.DryRun {
		if err := dirs.Prepare(false); err != nil {
			fmt.Fprintln(stdout, err)
			return 1
		}
		created, applyErr := linkapply.Apply(candidates, stdout)
		saveErr := history.Save(dirs.AppDataRoot, module.SourceDir, module.TargetDir, created)
		if saveErr != nil {
			fmt.Fprintf(stdout, "save link history: %v\n", saveErr)
		}
		if applyErr != nil || saveErr != nil {
			return 1
		}
		return 0
	}
	fmt.Fprintln(stdout, "Dry run. No links will be created. The following links would be created:")
	for _, candidate := range candidates {
		_, err := os.Lstat(candidate.Target)
		switch {
		case err == nil:
			fmt.Fprintf(stdout, "Item already exists: %s\n", candidate.Target)
		case os.IsNotExist(err):
			fmt.Fprintf(stdout, "Linking: '%s' '%s'\n", candidate.Source, candidate.Target)
		default:
			fmt.Fprintf(stdout, "inspect destination %q: %v\n", candidate.Target, err)
			return 1
		}
	}
	return 0
}
