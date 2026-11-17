package pathresolve

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ADIX7/alma2/internal/testutil"
)

func TestResolvePaths(t *testing.T) {
	f := testutil.NewFixture(t)
	documents := filepath.Join(f.Home, "Documents")
	r := Resolver{OS: runtime.GOOS, Home: f.Home, Documents: documents}
	for _, tt := range []struct {
		name, value, base, want string
		yaml                    bool
	}{
		{"home only", "~", f.Target, f.Home, true},
		{"home relative", "~/.config/fish", f.Target, filepath.Join(f.Home, ".config", "fish"), true},
		{"home anywhere bypasses base", "before~/after", f.Target, filepath.Clean("before" + f.Home + "/after"), true},
		{"documents", "%DOCUMENTS%/PowerShell", f.Target, filepath.Join(documents, "PowerShell"), true},
		{"relative", "shell/fish", f.Target, filepath.Join(f.Target, "shell", "fish"), true},
		{"absolute", filepath.ToSlash(filepath.Join(f.Root, "absolute")), f.Target, filepath.Join(f.Root, "absolute"), true},
		{"no base", "a/b", "", filepath.Join("a", "b"), true},
		{"unknown token literal", "%OTHER%/fish", f.Target, filepath.Join(f.Target, "%OTHER%", "fish"), true},
		{"non YAML backslash permitted", `a\b`, f.Target, filepath.Join(f.Target, `a\b`), false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.Resolve(tt.value, tt.base, tt.yaml)
			if err != nil || got != tt.want {
				t.Fatalf("Resolve(%q) = %q, %v; want %q", tt.value, got, err, tt.want)
			}
		})
	}
}

func TestResolveErrorsAndWindowsAppData(t *testing.T) {
	f := testutil.NewFixture(t)
	r := Resolver{OS: "linux", Home: f.Home, Documents: filepath.Join(f.Home, "Documents"), AppData: f.DataBase}
	for _, tt := range []struct {
		name, value string
		resolver    Resolver
	}{
		{"YAML backslash", `foo\bar`, r},
		{"YAML Windows path", `C:\Users\me`, r},
		{"non-Windows APPDATA", "%APPDATA%/fish", r},
		{"unavailable home", "~", Resolver{OS: "linux"}},
		{"unavailable documents", "%DOCUMENTS%", Resolver{OS: "linux", Home: f.Home}},
		{"missing Windows APPDATA", "%APPDATA%", Resolver{OS: "windows", Home: f.Home}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := tt.resolver.Resolve(tt.value, f.Target, true); err == nil {
				t.Fatalf("Resolve(%q) = %q, want error", tt.value, got)
			}
		})
	}
	windows := Resolver{OS: "windows", Home: f.Home, AppData: f.DataBase}
	got, err := windows.Resolve("%APPDATA%/fish", f.Target, true)
	if err != nil || got != filepath.Join(f.DataBase, "fish") {
		t.Fatalf("Windows APPDATA = %q, %v", got, err)
	}
	if strings.Contains(got, "%APPDATA%") {
		t.Fatal("APPDATA token not expanded")
	}
}
