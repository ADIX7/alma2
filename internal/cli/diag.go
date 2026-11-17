package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/ADIX7/alma2/internal/pathresolve"
	"github.com/ADIX7/alma2/internal/startup"
)

func runDiag(inv Invocation, stdout io.Writer) int {
	switch inv.Helper {
	case "":
		fmt.Fprintln(stdout, "No diagnostic helper specified.")
		return 1
	case "--help":
		fmt.Fprintln(stdout, "Usage: diag list | diag <diag-helper>")
		return 0
	case "list":
		fmt.Fprintln(stdout, "Available diagnostic helpers:")
		for _, name := range []string{"app-data-providers", "config-home-providers", "special-path-resolver"} {
			fmt.Fprintln(stdout, name)
		}
		return 0
	case "special-path-resolver":
		paths := pathresolve.Current()
		fmt.Fprintln(stdout, "Special path resolvers:")
		fmt.Fprintf(stdout, "~ => %s\n%%DOCUMENTS%% => %s\n", valueOrNull(paths.Home), valueOrNull(paths.Documents))
		appData := "<unsupported>"
		if paths.OS == "windows" {
			appData = valueOrNull(paths.AppData)
		}
		fmt.Fprintf(stdout, "%%APPDATA%% => %s\n", appData)
		return 0
	case "config-home-providers", "app-data-providers":
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(stdout, err)
			return 1
		}
		dirs := startup.Resolve(startup.CurrentInputs(cwd))
		providers := dirs.ConfigProviders
		if inv.Helper == "app-data-providers" {
			providers = dirs.AppDataProviders
		}
		fmt.Fprintf(stdout, "There are %d providers:\n", len(providers))
		for _, provider := range providers {
			fmt.Fprintf(stdout, "%s => %s\n", provider.Name, valueOrNull(provider.Path))
		}
		return 0
	default:
		fmt.Fprintf(stdout, "Diagnostic helper %s is not found.\n", inv.Helper)
		return 1
	}
}

func valueOrNull(value string) string {
	if value == "" {
		return "<null>"
	}
	return value
}
