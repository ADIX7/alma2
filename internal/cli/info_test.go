package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ADIX7/alma2/internal/testutil"
)

func TestGeneralInfoAndRepositoryAvailability(t *testing.T) {
	f := testutil.NewFixture(t)
	status, output := runFixtureCLI(t, f, "info")
	if status != 0 || !strings.Contains(output, "Alma ") || !strings.Contains(output, "AppData folder: "+filepath.Join(f.DataBase, "alma")) || !strings.Contains(output, "Platform is '") || !strings.Contains(output, "No repositories found") {
		t.Fatalf("no config info: %d %q", status, output)
	}
	f.WriteFile(t, "config/alma/repository.yml", fmt.Sprintf("repositories:\n  - name: present\n    repositoryPath: '%s'\n  - name: absent\n    repositoryPath: '%s'\n", filepath.ToSlash(f.Repository), filepath.ToSlash(filepath.Join(f.Root, "missing"))))
	status, output = runFixtureCLI(t, f, "info")
	if status != 0 || !strings.Contains(output, "present\n") || !strings.Contains(output, "absent (containing folder not exists") {
		t.Fatalf("repository availability: %d %q", status, output)
	}
	if _, err := os.Lstat(filepath.Join(f.DataBase, "alma")); !os.IsNotExist(err) {
		t.Fatalf("info wrote app-data: %v", err)
	}
}

func TestRepositoryInfoMissingSourceAndRelativePath(t *testing.T) {
	f := testutil.NewFixture(t)
	f.WriteFile(t, "config/alma/repository.yml", "repositories: [{name: relative, repositoryPath: '../repository'}, {name: missing, repositoryPath: '../absent'}]\n")
	f.WriteFile(t, "repository/mod/.alma.yml", "platforms: []\n")
	status, output := runFixtureCLI(t, f, "info", "relative")
	if status != 0 || !strings.Contains(output, "contains 1 modules") {
		t.Fatalf("relative source: %d %q", status, output)
	}
	status, output = runFixtureCLI(t, f, "info", "missing")
	if status == 0 || !strings.Contains(output, "repository source") {
		t.Fatalf("missing source: %d %q", status, output)
	}
}

func TestRepositoryInfoSortedCountsAndCapabilities(t *testing.T) {
	f := testutil.NewFixture(t)
	f.WriteFile(t, "config/alma/repository.yml", fmt.Sprintf("repositories: [{name: fixtures, repositoryPath: '%s'}]\n", filepath.ToSlash(f.Repository)))
	for _, args := range [][]string{{"info", "fixtures"}, {"info", "missing"}} {
		status, output := runFixtureCLI(t, f, args...)
		if args[1] == "fixtures" && (status != 0 || !strings.Contains(output, "contains 0 modules")) {
			t.Fatalf("empty: %d %q", status, output)
		}
		if args[1] == "missing" && (status == 0 || !strings.Contains(output, "unknown repository")) {
			t.Fatalf("unknown: %d %q", status, output)
		}
	}
	f.WriteFile(t, "repository/zeta/.alma.yml", "platforms: [{platform: default, links: [{source: a, target: b}, {source: c, target: d}], install: 'echo harmless', configure: 'echo harmless'}]\n")
	f.WriteFile(t, "repository/alpha/.alma.yml", "platforms: [{platform: default, links: [{source: x, target: y}]}]\n")
	f.WriteFile(t, "repository/unavailable/.alma.yml", "platforms: [{platform: never-supported}]\n")
	f.WriteFile(t, "repository/zeta/deeper/.alma.yml", "platforms: []\n")
	status, output := runFixtureCLI(t, f, "info", "fixtures")
	if status != 0 || !strings.Contains(output, "contains 3 modules") || !strings.Contains(output, "alpha  1 link(s)") || !strings.Contains(output, "zeta  2 link(s)    [Install]    [Configure]") || !strings.Contains(output, "unavailable  [Unavailable]") {
		t.Fatalf("modules: %d %q", status, output)
	}
	if !(strings.Index(output, "alpha  ") < strings.Index(output, "unavailable  ") && strings.Index(output, "unavailable  ") < strings.Index(output, "zeta  ")) {
		t.Fatalf("not sorted: %q", output)
	}
}
