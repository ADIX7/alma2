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
	// Windows short directory names such as RUNNER~1 contain a literal tilde.
	// Expand only home tokens; preserve the short-name suffix in YAML paths.
	var expanded strings.Builder
	hasHome := false
	for i := 0; i < len(value); i++ {
		if value[i] != '~' || (r.OS == "windows" && i > 0 && i+1 < len(value) &&
			isAlphaNumeric(value[i-1]) && value[i+1] >= '0' && value[i+1] <= '9') {
			expanded.WriteByte(value[i])
			continue
		}
		if r.Home == "" {
			return "", fmt.Errorf("home directory unavailable")
		}
		hasHome = true
		expanded.WriteString(r.Home)
	}
	value = expanded.String()
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

// ResolveDestination resolves a YAML destination without letting a relative
// path implicitly refer to the process working directory.
func (r Resolver) ResolveDestination(value, base string) (string, error) {
	resolved, err := r.Resolve(value, base, true)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(resolved) {
		if base == "" {
			return "", fmt.Errorf("destination %q must be absolute when no target directory is configured", value)
		}
		return "", fmt.Errorf("destination %q did not resolve to an absolute path", value)
	}
	return resolved, nil
}

func isAlphaNumeric(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}
