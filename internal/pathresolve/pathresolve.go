// Package pathresolve expands Alma path expressions without touching the filesystem.
package pathresolve

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Resolver supplies platform-specific special directories. OS can be injected
// in tests; production uses runtime.GOOS.
type Resolver struct {
	OS, Home, Documents, AppData string
}

func Current() Resolver {
	home, _ := os.UserHomeDir()
	documents := ""
	if home != "" {
		documents = filepath.Join(home, "Documents")
	}
	appData := ""
	if runtime.GOOS == "windows" {
		appData = os.Getenv("APPDATA")
	}
	return Resolver{OS: runtime.GOOS, Home: home, Documents: documents, AppData: appData}
}

// Resolve replaces literal tokens anywhere in a path. A path containing '~'
// bypasses base combination, matching the legacy home shortcut. YAML values
// must use forward slashes; command strings should never be passed here.
func (r Resolver) Resolve(value, base string, fromYAML bool) (string, error) {
	if fromYAML && strings.Contains(value, `\`) {
		return "", fmt.Errorf("YAML path must use / separators: %q", value)
	}
	hasHome := strings.Contains(value, "~")
	if hasHome {
		if r.Home == "" {
			return "", fmt.Errorf("home directory unavailable")
		}
		value = strings.ReplaceAll(value, "~", r.Home)
	}
	if strings.Contains(value, "%DOCUMENTS%") {
		if r.Documents == "" {
			return "", fmt.Errorf("Documents directory unavailable")
		}
		value = strings.ReplaceAll(value, "%DOCUMENTS%", r.Documents)
	}
	if strings.Contains(value, "%APPDATA%") {
		if r.OS != "windows" || r.AppData == "" {
			return "", fmt.Errorf("%%APPDATA%% is unsupported on %s", r.OS)
		}
		value = strings.ReplaceAll(value, "%APPDATA%", r.AppData)
	}
	value = filepath.FromSlash(value)
	if !hasHome && !filepath.IsAbs(value) && base != "" {
		value = filepath.Join(base, value)
	}
	return filepath.Clean(value), nil
}
