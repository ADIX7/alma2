package operation

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ADIX7/alma2/internal/config"
	"github.com/ADIX7/alma2/internal/pathresolve"
	"github.com/ADIX7/alma2/internal/startup"
	"github.com/ADIX7/alma2/internal/testutil"
)

func str(value string) *string { return &value }

func setup(t *testing.T) (*testutil.Fixture, startup.Directories, pathresolve.Resolver) {
	t.Helper()
	f := testutil.NewFixture(t)
	return f, startup.Directories{WorkDir: f.Work}, pathresolve.Resolver{OS: runtime.GOOS, Home: f.Home, Documents: filepath.Join(f.Home, "Documents")}
}

func TestResolveRepositoryDefaultsAndNames(t *testing.T) {
	f, dirs, paths := setup(t)
	entries := []config.Repository{
		{Name: "explicit", RepositoryPath: str("../repository"), LinkPath: str("../target")},
		{Name: "partial", LinkPath: str("~/.links")},
		{Name: "only-source", RepositoryPath: str("~/sources")},
	}
	for _, tc := range []struct {
		name, source, target string
		present              bool
	}{
		{"", f.Work, f.Root, false},
		{"explicit", f.Repository, f.Target, true},
		{"partial", f.Work, filepath.Join(f.Home, ".links"), false},
		{"only-source", filepath.Join(f.Home, "sources"), f.Root, true},
	} {
		got, err := ResolveRepository(dirs, paths, entries, tc.name)
		if err != nil || got.SourceRoot != tc.source || got.TargetRoot != tc.target || got.RepositoryPathPresent != tc.present {
			t.Fatalf("repo %q = %+v, %v; want source=%q target=%q present=%v", tc.name, got, err, tc.source, tc.target, tc.present)
		}
	}
	if _, err := ResolveRepository(dirs, paths, entries, "EXPLICIT"); err == nil || !strings.Contains(err.Error(), "unknown repository") {
		t.Fatalf("unknown explicit repository: %v", err)
	}
}

func TestResolveModuleTwoFormsAndTargets(t *testing.T) {
	f, dirs, paths := setup(t)
	f.WriteFile(t, "work/shell/fish/file", "payload")
	f.WriteFile(t, "repository/shell/fish/.alma.yml", `platforms:
  - platform: default
    target: 'module-target'
  - platform: linux
    target: '~/.config/fish'
`)
	entries := []config.Repository{{Name: "dotfiles", RepositoryPath: str("../repository"), LinkPath: str("../target")}}
	defaultModule, err := ResolveModule(dirs, paths, entries, "", "shell/fish", "linux-x64")
	if err != nil || defaultModule.SourceDir != filepath.Join(f.Work, "shell", "fish") || defaultModule.TargetDir != f.Root || !defaultModule.Settings.Available || defaultModule.Settings.ExcludeReadme {
		t.Fatalf("one-position fallback: %+v, %v", defaultModule, err)
	}
	configured, err := ResolveModule(dirs, paths, entries, "dotfiles", "shell/fish", "windows-x64")
	if err != nil || configured.SourceDir != filepath.Join(f.Repository, "shell", "fish") || configured.TargetDir != filepath.Join(f.Target, "module-target") || !configured.Settings.ExcludeReadme {
		t.Fatalf("two-position config: %+v, %v", configured, err)
	}
	linux, err := ResolveModule(dirs, paths, entries, "dotfiles", "shell/fish", "linux-x64")
	if err != nil || linux.TargetDir != filepath.Join(f.Home, ".config", "fish") {
		t.Fatalf("absolute/home module override: %+v, %v", linux, err)
	}
	if _, err := os.Stat(filepath.Join(f.Target, "module-target")); !os.IsNotExist(err) {
		t.Fatalf("resolution created target: %v", err)
	}
	root, err := ResolveModule(dirs, paths, entries, "dotfiles", ".", "linux-x64")
	if err != nil || root.SourceDir != f.Repository || !root.Settings.Available {
		t.Fatalf("repository-root module: %+v, %v", root, err)
	}
}

func TestResolveModuleErrorsAndUnavailability(t *testing.T) {
	f, dirs, paths := setup(t)
	entries := []config.Repository{{Name: "dotfiles", RepositoryPath: str("../repository")}}
	f.WriteFile(t, "repository/configured/.alma.yml", "platforms: [{platform: windows}]\n")
	got, err := ResolveModule(dirs, paths, entries, "dotfiles", "configured", "linux-x64")
	if err != nil || got.Settings.Available || got.TargetDir != f.Root {
		t.Fatalf("unavailable module: %+v, %v", got, err)
	}
	for _, tc := range []struct{ repo, module, message string }{
		{"unknown", "configured", "unknown repository"},
		{"dotfiles", "missing", "module directory"},
		{"dotfiles", "../target", "unsafe module"},
	} {
		if _, err := ResolveModule(dirs, paths, entries, tc.repo, tc.module, "linux-x64"); err == nil || !strings.Contains(err.Error(), tc.message) {
			t.Fatalf("repo %q module %q: %v; want %q", tc.repo, tc.module, err, tc.message)
		}
	}
	if err := os.WriteFile(filepath.Join(f.Repository, "file"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveModule(dirs, paths, entries, "dotfiles", "file", "linux-x64"); err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("file module: %v", err)
	}
	f.WriteFile(t, "repository/broken/.alma.yml", "platforms: null\n")
	if _, err := ResolveModule(dirs, paths, entries, "dotfiles", "broken", "linux-x64"); err == nil || !strings.Contains(err.Error(), "null") {
		t.Fatalf("invalid module YAML: %v", err)
	}
	if _, err := ResolveModule(startup.Directories{WorkDir: filepath.Join(f.Root, "absent")}, paths, nil, "", "mod", "linux-x64"); err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing source directory: %v", err)
	}
}

func TestModuleBoundaryAndSymlinks(t *testing.T) {
	f, dirs, paths := setup(t)
	f.WriteFile(t, "repository/safe/nested/file", "payload")
	entries := []config.Repository{{Name: "dotfiles", RepositoryPath: str("../repository")}}
	for _, name := range []string{"", "..", "../repository/safe", "safe/../../../target", "/tmp/outside", f.Target, `C:/outside`, `safe\nested`} {
		if _, err := ResolveModule(dirs, paths, entries, "dotfiles", name, "linux-x64"); err == nil || !strings.Contains(err.Error(), "unsafe module name") {
			t.Fatalf("unsafe module %q accepted: %v", name, err)
		}
	}
	if got, err := ResolveModule(dirs, paths, entries, "dotfiles", "safe/nested", "linux-x64"); err != nil || got.SourceDir != filepath.Join(f.Repository, "safe", "nested") {
		t.Fatalf("valid nested module: %+v, %v", got, err)
	}
	// A non-escaping .. component is allowed after lexical normalization.
	if got, err := ResolveModule(dirs, paths, entries, "dotfiles", "safe/nested/../nested", "linux-x64"); err != nil || got.SourceDir != filepath.Join(f.Repository, "safe", "nested") {
		t.Fatalf("non-escaping ..: %+v, %v", got, err)
	}
	if err := os.Symlink(f.Target, filepath.Join(f.Repository, "external")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	if err := os.Symlink(filepath.Join(f.Repository, "safe"), filepath.Join(f.Repository, "alias")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	for _, name := range []string{"external", "alias/nested"} {
		if _, err := ResolveModule(dirs, paths, entries, "dotfiles", name, "linux-x64"); err == nil || !strings.Contains(err.Error(), "symlinked directory") {
			t.Fatalf("symlink module %q: %v", name, err)
		}
	}
	if err := os.Symlink(f.Repository, filepath.Join(f.Root, "repo-alias")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	entries[0].RepositoryPath = str("../repo-alias")
	if _, err := ResolveModule(dirs, paths, entries, "dotfiles", ".", "linux-x64"); err == nil || !strings.Contains(err.Error(), "symlinked directory") {
		t.Fatalf("symlinked repository root module: %v", err)
	}
}
