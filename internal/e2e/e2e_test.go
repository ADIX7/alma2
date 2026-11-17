package e2e

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ADIX7/alma2/internal/history"
	"github.com/ADIX7/alma2/internal/testutil"
	"gopkg.in/yaml.v3"
)

func buildFixtureCLI(t *testing.T, f *testutil.Fixture) string {
	t.Helper()
	name := "alma"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(f.Bin, name)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/alma")
	command.Dir = filepath.Join("..", "..")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build fixture CLI: %v: %s", err, output)
	}
	return binary
}

func fixtureFile(t *testing.T, f *testutil.Fixture, from, to string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", from))
	if err != nil {
		t.Fatal(err)
	}
	f.WriteFile(t, to, string(data))
}

func invoke(t *testing.T, f *testutil.Fixture, executable string, args ...string) string {
	t.Helper()
	result, err := f.RunCLI(executable, args...)
	if err != nil || result.ExitCode != 0 {
		t.Fatalf("alma %v: %+v, %v", args, result, err)
	}
	return result.Stdout
}

func TestYAMLToDryRunLinkHistoryAndRepeat(t *testing.T) {
	f := testutil.NewFixture(t)
	source := f.WriteFile(t, "repository/mod/unmapped", "unmapped")
	fixtureFile(t, f, "repository.yml", "config/alma/repository.yml")
	fixtureFile(t, f, "mod/.alma.yml", "repository/mod/.alma.yml")
	f.WriteFile(t, "repository/mod/.alma-config.json", `{ "legacy": true }`)
	f.WriteFile(t, "repository/mod/README.md", "# literal fixture readme")
	f.WriteFile(t, "repository/mod/editor/config", "editor")
	f.WriteFile(t, "repository/mod/other/config", "other")
	f.WriteFile(t, "repository/mod/folder/child", "folder")
	f.WriteFile(t, "repository/mod/scratch/skip", "skip")
	probe := filepath.Join(f.Target, "probe")
	if err := os.Symlink(source, probe); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	if err := os.Remove(probe); err != nil {
		t.Fatal(err)
	}
	binary := buildFixtureCLI(t, f)

	if output := invoke(t, f, binary, "ls"); output != "Repositories\nfixtures\n" {
		t.Fatalf("repositories: %q", output)
	}
	if output := invoke(t, f, binary, "ls", "fixtures"); output != "Modules in repository 'fixtures'\nmod\n" {
		t.Fatalf("modules: %q", output)
	}
	if output := invoke(t, f, binary, "info", "fixtures"); !strings.Contains(output, "mod  2 link(s)") {
		t.Fatalf("mapping count: %q", output)
	}
	if output := invoke(t, f, binary, "readme", "fixtures", "mod"); output != "# literal fixture readme" {
		t.Fatalf("literal readme: %q", output)
	}

	preview := invoke(t, f, binary, "link", "fixtures", "mod", "--dry-run")
	for _, name := range []string{"renamed/config", "mapped-folder", "other/config", "unmapped"} {
		if !strings.Contains(preview, filepath.Join(f.Target, filepath.FromSlash(name))) {
			t.Fatalf("missing planned target %s: %q", name, preview)
		}
	}
	for _, name := range []string{"README.md", "scratch/skip", ".alma.yml", ".alma-config.json"} {
		if strings.Contains(preview, filepath.Join(f.Repository, "mod", filepath.FromSlash(name))) {
			t.Fatalf("excluded source %s planned: %q", name, preview)
		}
	}
	if _, err := os.Lstat(filepath.Join(f.DataBase, "alma")); !os.IsNotExist(err) {
		t.Fatalf("dry run wrote app data: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(f.Target, "renamed")); !os.IsNotExist(err) {
		t.Fatalf("dry run wrote target parent: %v", err)
	}

	invoke(t, f, binary, "link", "fixtures", "mod")
	for _, name := range []string{"renamed/config", "mapped-folder", "other/config", "unmapped"} {
		target := filepath.Join(f.Target, filepath.FromSlash(name))
		info, err := os.Lstat(target)
		if err != nil || info.Mode()&os.ModeSymlink == 0 {
			t.Fatalf("missing link %s: %v", name, err)
		}
	}
	for _, name := range []string{"README.md", "scratch/skip", ".alma.yml", ".alma-config.json", "folder/child"} {
		if _, err := os.Lstat(filepath.Join(f.Target, filepath.FromSlash(name))); !os.IsNotExist(err) {
			t.Fatalf("unexpected target %s: %v", name, err)
		}
	}
	if linked, err := os.Readlink(filepath.Join(f.Target, "mapped-folder")); err != nil || linked != filepath.Join(f.Repository, "mod", "folder") {
		t.Fatalf("directory mapping: %q, %v", linked, err)
	}
	recordFile, err := history.File(filepath.Join(f.DataBase, "alma"), filepath.Join(f.Repository, "mod"), f.Target)
	if err != nil {
		t.Fatal(err)
	}
	readHistory := func() history.Record {
		t.Helper()
		contents, err := os.ReadFile(recordFile)
		if err != nil {
			t.Fatal(err)
		}
		var record history.Record
		if err := yaml.Unmarshal(contents, &record); err != nil {
			t.Fatal(err)
		}
		return record
	}
	if record := readHistory(); len(record.Destinations) != 4 {
		t.Fatalf("history: %+v", record)
	}
	if output := invoke(t, f, binary, "link", "fixtures", "mod"); strings.Count(output, "Item already exists:") != 4 {
		t.Fatalf("repeat conflicts: %q", output)
	}
	if record := readHistory(); len(record.Destinations) != 4 {
		t.Fatalf("repeat history: %+v", record)
	}
}
