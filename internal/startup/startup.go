// Package startup selects Alma's working, configuration, and application-data directories.
package startup

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Inputs allows environment and OS directory providers to be controlled in
// tests. Empty values represent unavailable providers.
type Inputs struct {
	Env           func(string) string
	InvocationDir string
	Home          string
	LocalData     string
}

// Provider records the value of a directory provider for diagnostics.
type Provider struct {
	Name, Path string
}

// Directories are selected without creating anything. ConfigRoot may be empty
// when no provider is available. AppDataRoot always has a working-dir fallback.
type Directories struct {
	WorkDir, ConfigRoot, AppDataRoot  string
	ConfigProviders, AppDataProviders []Provider
	Warnings                          []string
}

func existingDir(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func choose(providers []Provider) string {
	for _, provider := range providers {
		if existingDir(provider.Path) {
			return provider.Path
		}
	}
	return providers[len(providers)-1].Path
}

// Resolve selects the first existing provider in each documented precedence
// list. A missing final fallback is retained; no paths are created here.
func Resolve(in Inputs) Directories {
	env := in.Env
	if env == nil {
		env = os.Getenv
	}
	result := Directories{WorkDir: in.InvocationDir}
	for _, name := range []string{"ALMA_WORKDIR", "WORKDIR"} {
		path := env(name)
		if path == "" {
			continue
		}
		if existingDir(path) {
			result.WorkDir = path
			break
		}
		result.Warnings = append(result.Warnings, fmt.Sprintf("%s is not an existing directory: %s", name, path))
	}
	userConfig := ""
	if in.Home != "" {
		userConfig = filepath.Join(in.Home, ".config")
	}
	result.ConfigProviders = []Provider{
		{"ALMA_CONFIG", env("ALMA_CONFIG")},
		{"XDG_CONFIG_HOME", env("XDG_CONFIG_HOME")},
		{"user config", userConfig},
		{"ALMA_CONFIG_FALLBACK", env("ALMA_CONFIG_FALLBACK")},
	}
	configBase := choose(result.ConfigProviders)
	if configBase != "" {
		result.ConfigRoot = filepath.Join(configBase, "alma")
	}
	result.AppDataProviders = []Provider{
		{"ALMA_APP_DATA", env("ALMA_APP_DATA")},
		{"local app data", in.LocalData},
		{"ALMA_APP_DATA_FALLBACK", env("ALMA_APP_DATA_FALLBACK")},
	}
	dataBase := choose(result.AppDataProviders)
	if dataBase == "" {
		dataBase = result.WorkDir
	}
	result.AppDataRoot = filepath.Join(dataBase, "alma")
	return result
}

// Prepare creates application data only for a real operation. Dry runs remain
// read-only, including when a fallback directory does not yet exist.
func (d Directories) Prepare(dryRun bool) error {
	if dryRun {
		return nil
	}
	return os.MkdirAll(d.AppDataRoot, 0700)
}

// CurrentInputs obtains OS defaults; callers supply the invocation directory.
// A local-data value is deliberately separate from the configuration provider.
func CurrentInputs(invocationDir string) Inputs {
	home, _ := os.UserHomeDir()
	localData := os.Getenv("XDG_DATA_HOME")
	if runtime.GOOS == "windows" {
		localData = os.Getenv("LOCALAPPDATA")
	} else if localData == "" && home != "" {
		if runtime.GOOS == "darwin" {
			localData = filepath.Join(home, "Library", "Application Support")
		} else {
			localData = filepath.Join(home, ".local", "share")
		}
	}
	return Inputs{Env: os.Getenv, InvocationDir: invocationDir, Home: home, LocalData: localData}
}
