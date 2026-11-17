package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ADIX7/alma2/internal/testutil"
)

func runListFixture(t *testing.T, f *testutil.Fixture, args ...string) (int, string) {
	t.Helper()
	t.Setenv("ALMA_CONFIG", f.ConfigBase)
	t.Setenv("ALMA_WORKDIR", f.Work)
	t.Setenv("HOME", f.Home)
	t.Setenv("USERPROFILE", f.Home)
	var output bytes.Buffer
	status := Run(args, &output)
	return status, output.String()
}

func TestListRepositoriesAndModules(t *testing.T) {
	f := testutil.NewFixture(t)
	if status, output := runListFixture(t, f, "ls"); status != 0 || output != "Repositories\n" {
		t.Fatalf("no config: %d %q", status, output)
	}
	f.WriteFile(t, "config/alma/repository.yml", fmt.Sprintf("repositories:\n  - name: dotfiles\n    repositoryPath: '%s'\n  - name: other\n    repositoryPath: '%s'\n", filepath.ToSlash(f.Repository), filepath.ToSlash(f.Work)))
	f.WriteFile(t, "repository/.alma.yml", "platforms: []\n")
	f.WriteFile(t, "repository/editor/.alma.yml", "platforms: [{platform: windows}]\n")
	f.WriteFile(t, "repository/editor/nested/.alma.yml", "platforms: []\n")
	f.WriteFile(t, "repository/.hidden/.alma.yml", "platforms: []\n")
	f.WriteFile(t, "repository/legacy/.alma-config.json", `{"platforms":[]}`)
	f.WriteFile(t, "repository/legacy/.alma.yaml", "platforms: []\n")
	f.WriteFile(t, "repository/regular/file", "payload")
	if err := os.Symlink(filepath.Join(f.Repository, "editor"), filepath.Join(f.Repository, "symlinked")); err != nil {
		t.Logf("symlink traversal assertion unavailable: %v", err)
	}
	if status, output := runListFixture(t, f, "ls", "extra"); status != 1 || !strings.Contains(output, "unknown repository") {
		t.Fatalf("unknown name: %d %q", status, output)
	}
	if status, output := runListFixture(t, f, "ls"); status != 0 || output != "Repositories\ndotfiles\nother\n" {
		t.Fatalf("repository order: %d %q", status, output)
	}
	if status, output := runListFixture(t, f, "ls", "dotfiles", "ignored"); status != 0 || output != "Modules in repository 'dotfiles'\n.hidden\neditor\n" {
		t.Fatalf("module discovery: %d %q", status, output)
	}
	if status, output := runListFixture(t, f, "ls", "other"); status != 0 || output != "Modules in repository 'other'\n" {
		t.Fatalf("empty repository: %d %q", status, output)
	}
	if _, err := os.Stat(filepath.Join(f.DataBase, "alma")); !os.IsNotExist(err) {
		t.Fatalf("ls created app-data: %v", err)
	}
}

func TestListErrorsAndPathResolution(t *testing.T) {
	f := testutil.NewFixture(t)
	f.WriteFile(t, "config/alma/repository.yml", "repositories:\n  - name: without-path\n  - name: tilde\n    repositoryPath: '~/modules'\n  - name: relative\n    repositoryPath: '../repository'\n  - name: absent\n    repositoryPath: '../missing'\n")
	f.WriteFile(t, "home/modules/.alma.yml", "platforms: []\n")
	f.WriteFile(t, "home/modules/fish/.alma.yml", "platforms: []\n")
	f.WriteFile(t, "repository/invalid/.alma.yml", "platforms: null\n")
	for _, tc := range []struct{ name, message string }{
		{"WITHOUT-PATH", "unknown repository"},
		{"without-path", "No repository path"},
		{"absent", "repository source"},
		{"relative", "null"},
	} {
		status, output := runListFixture(t, f, "ls", tc.name)
		if status == 0 || !strings.Contains(output, tc.message) {
			t.Fatalf("ls %q: %d %q; want %q", tc.name, status, output, tc.message)
		}
	}
	if status, output := runListFixture(t, f, "ls", "tilde"); status != 0 || output != "Modules in repository 'tilde'\nfish\n" {
		t.Fatalf("resolved tilde path: %d %q", status, output)
	}
	f.WriteFile(t, "config/alma/repository.yml", "repositories: null\n")
	if status, output := runListFixture(t, f, "ls"); status == 0 || !strings.Contains(output, "null") {
		t.Fatalf("invalid repository: %d %q", status, output)
	}
}
