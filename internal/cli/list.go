package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/ADIX7/alma2/internal/config"
	"github.com/ADIX7/alma2/internal/operation"
	"github.com/ADIX7/alma2/internal/pathresolve"
	"github.com/ADIX7/alma2/internal/startup"
)

func runList(inv Invocation, stdout io.Writer) int {
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
	if inv.Repository == "" {
		fmt.Fprintln(stdout, "Repositories")
		for _, repository := range repositories {
			fmt.Fprintln(stdout, repository.Name)
		}
		return 0
	}
	repo, err := operation.ResolveRepository(dirs, pathresolve.Current(), repositories, inv.Repository)
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}
	if !repo.RepositoryPathPresent {
		fmt.Fprintf(stdout, "No repository path is specified in repository settings '%s'\n", inv.Repository)
		return 1
	}
	modules, err := operation.DiscoverModules(repo.SourceRoot)
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}
	fmt.Fprintf(stdout, "Modules in repository '%s'\n", inv.Repository)
	for _, module := range modules {
		fmt.Fprintln(stdout, module)
	}
	return 0
}
