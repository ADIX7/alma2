package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/ADIX7/alma2/internal/config"
	"github.com/ADIX7/alma2/internal/operation"
	"github.com/ADIX7/alma2/internal/pathresolve"
	"github.com/ADIX7/alma2/internal/platform"
	"github.com/ADIX7/alma2/internal/shell"
	"github.com/ADIX7/alma2/internal/startup"
)

func runConfiguredCommand(inv Invocation, stdout io.Writer) int {
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
	if !module.Settings.Available {
		fmt.Fprintf(stdout, "Module '%s' is unavailable on this platform.\n", inv.Module)
		return 1
	}
	configured, err := config.LoadModule(module.SourceDir)
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}
	if configured == nil {
		fmt.Fprintf(stdout, "No configuration is found for module '%s'.\n", inv.Module)
		return 0
	}
	var command *string
	if inv.Command == "install" {
		command = module.Settings.Install
	} else {
		command = module.Settings.Configure
	}
	if command == nil {
		fmt.Fprintf(stdout, "No %s command for module '%s'.\n", inv.Command, inv.Module)
		return 0
	}
	fmt.Fprintln(stdout, *command)
	if err := shell.Execute(context.Background(), runtime.GOOS, *command, dirs.WorkDir, shell.ProcessRunner{}, stdout, os.Stderr); err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}
	return 0
}
