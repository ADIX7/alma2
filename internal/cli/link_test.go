package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ADIX7/alma2/internal/history"
	"github.com/ADIX7/alma2/internal/testutil"
	"gopkg.in/yaml.v3"
)

func linkFixture(t *testing.T) *testutil.Fixture {
	t.Helper()
	f := testutil.NewFixture(t)
	f.WriteFile(t, "config/alma/repository.yml", fmt.Sprintf("repositories:\n  - name: fixtures\n    repositoryPath: '%s'\n    linkPath: '%s'\n", filepath.ToSlash(f.Repository), filepath.ToSlash(f.Target)))
	return f
}

func TestLinkDryRunProposalsAndConflicts(t *testing.T) {
	f := linkFixture(t)
	f.WriteFile(t, "repository/mod/.alma.yml", "platforms: [{platform: default, links: [{source: move, target: renamed/move}]}]\n")
	f.WriteFile(t, "repository/mod/move", "renamed")
	f.WriteFile(t, "repository/mod/nested/file", "nested")
	f.WriteFile(t, "repository/mod/README.md", "excluded")
	f.WriteFile(t, "repository/mod/.alma-config.json", `{}`)
	f.WriteFile(t, "repository/mod/existing-file", "file")
	f.WriteFile(t, "repository/mod/existing-dir", "directory")
	f.WriteFile(t, "repository/mod/existing-link", "symlink")
	f.WriteFile(t, "repository/mod/broken-link", "broken")
	f.WriteFile(t, "target/existing-file", "keep")
	if err := os.Mkdir(filepath.Join(f.Target, "existing-dir"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(f.Repository, "mod", "existing-link"), filepath.Join(f.Target, "existing-link")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	if err := os.Symlink(filepath.Join(f.Target, "absent"), filepath.Join(f.Target, "broken-link")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	for _, flag := range []string{"-d", "--dry-run"} {
		status, output := runFixtureCLI(t, f, "link", "fixtures", "mod", flag, "ignored")
		if status != 0 || !strings.Contains(output, "Dry run.") || !strings.Contains(output, "Linking: '") || !strings.Contains(output, filepath.Join(f.Target, "renamed", "move")) || !strings.Contains(output, filepath.Join(f.Target, "nested", "file")) {
			t.Fatalf("dry run %s: %d %q", flag, status, output)
		}
		for _, name := range []string{"existing-file", "existing-dir", "existing-link", "broken-link"} {
			if !strings.Contains(output, "Item already exists: "+filepath.Join(f.Target, name)) {
				t.Fatalf("missing conflict %s in %q", name, output)
			}
		}
		for _, name := range []string{".alma.yml", ".alma-config.json", "README.md"} {
			if strings.Contains(output, filepath.Join(f.Repository, "mod", name)) {
				t.Fatalf("protected payload %q proposed: %q", name, output)
			}
		}
		for _, path := range []string{filepath.Join(f.DataBase, "alma"), filepath.Join(f.Target, "renamed"), filepath.Join(f.Target, "nested")} {
			if _, err := os.Lstat(path); !os.IsNotExist(err) {
				t.Fatalf("dry run wrote %q: %v", path, err)
			}
		}
	}
	content, err := os.ReadFile(filepath.Join(f.Target, "existing-file"))
	if err != nil || string(content) != "keep" {
		t.Fatalf("conflict changed: %q, %v", content, err)
	}
}

func TestLinkDryRunFailuresAreReadOnly(t *testing.T) {
	f := linkFixture(t)
	f.WriteFile(t, "repository/mod/.alma.yml", "platforms: [{platform: default, links: [{source: a, target: same}, {source: b, target: same}]}]\n")
	f.WriteFile(t, "repository/mod/a", "a")
	f.WriteFile(t, "repository/mod/b", "b")
	for _, tc := range []struct {
		args    []string
		message string
	}{
		{[]string{"link", "fixtures", "mod", "-d"}, "duplicate destination"},
		{[]string{"link", "fixtures", "absent", "-d"}, "module directory"},
		{[]string{"link", "absent", "mod", "-d"}, "unknown repository"},
		{[]string{"link", "fixtures", "../mod", "-d"}, "unsafe module"},
	} {
		status, output := runFixtureCLI(t, f, tc.args...)
		if status == 0 || !strings.Contains(output, tc.message) {
			t.Fatalf("%v: %d %q; want %q", tc.args, status, output, tc.message)
		}
		if _, err := os.Lstat(filepath.Join(f.DataBase, "alma")); !os.IsNotExist(err) {
			t.Fatalf("failed dry run wrote app data: %v", err)
		}
	}
	if _, err := os.Lstat(filepath.Join(f.Target, "same")); !os.IsNotExist(err) {
		t.Fatalf("collision wrote destination: %v", err)
	}
}

func TestLinkDryRunWorkingDirectoryAndUnavailableModule(t *testing.T) {
	f := linkFixture(t)
	f.WriteFile(t, "work/plain/file", "payload")
	status, output := runFixtureCLI(t, f, "link", "plain", "-d")
	if status != 0 || !strings.Contains(output, filepath.Join(f.Work, "plain", "file")) || !strings.Contains(output, filepath.Join(f.Root, "file")) {
		t.Fatalf("one-position fallback: %d %q", status, output)
	}
	f.WriteFile(t, "repository/unavailable/.alma.yml", "platforms: [{platform: never-supported}]\n")
	status, output = runFixtureCLI(t, f, "link", "fixtures", "unavailable", "-d")
	if status == 0 || !strings.Contains(output, "unavailable") {
		t.Fatalf("unavailable module was planned: %d %q", status, output)
	}
	if _, err := os.Lstat(filepath.Join(f.DataBase, "alma")); !os.IsNotExist(err) {
		t.Fatalf("dry run wrote app data: %v", err)
	}
}

func TestLinkRealCollisionAndPartialFailure(t *testing.T) {
	f := linkFixture(t)
	f.WriteFile(t, "repository/mod/.alma.yml", "platforms: [{platform: default, links: [{source: a, target: same}, {source: b, target: same}]}]\n")
	f.WriteFile(t, "repository/mod/a", "a")
	f.WriteFile(t, "repository/mod/b", "b")
	status, output := runFixtureCLI(t, f, "link", "fixtures", "mod")
	if status == 0 || !strings.Contains(output, "duplicate destination") {
		t.Fatalf("collision: %d %q", status, output)
	}
	if _, err := os.Lstat(filepath.Join(f.DataBase, "alma")); !os.IsNotExist(err) {
		t.Fatalf("collision created app data: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(f.Target, "same")); !os.IsNotExist(err) {
		t.Fatalf("collision created target: %v", err)
	}
	f.WriteFile(t, "repository/mod/.alma.yml", "platforms: [{platform: default, links: [{source: a, target: blocker/failed}, {source: b, target: good}]}]\n")
	f.WriteFile(t, "target/blocker", "file")
	status, output = runFixtureCLI(t, f, "link", "fixtures", "mod")
	if status == 0 || !strings.Contains(output, "create link") || !strings.Contains(output, "Linking:") {
		t.Fatalf("partial run: %d %q", status, output)
	}
	info, err := os.Lstat(filepath.Join(f.Target, "good"))
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("partial success not retained: %v", err)
	}
	record := loadCLIHistory(t, f, "mod")
	if len(record.Destinations) != 1 || record.Destinations[0] != filepath.Join(f.Target, "good") {
		t.Fatalf("partial history: %+v", record)
	}
}

func TestLinkUnconfiguredAndRealWrites(t *testing.T) {
	f := linkFixture(t)
	f.WriteFile(t, "repository/plain/README.md", "plain")
	f.WriteFile(t, "repository/plain/keep", "source")
	f.WriteFile(t, "target/keep", "pre-existing")
	f.WriteFile(t, "repository/plain/.alma-config.json", `{}`)
	status, output := runFixtureCLI(t, f, "link", "fixtures", "plain", "-d")
	if status != 0 || !strings.Contains(output, filepath.Join(f.Repository, "plain", "README.md")) || strings.Contains(output, filepath.Join(f.Repository, "plain", ".alma-config.json")) {
		t.Fatalf("unconfigured: %d %q", status, output)
	}
	probe := filepath.Join(f.Target, "probe")
	if err := os.Symlink(filepath.Join(f.Repository, "plain", "README.md"), probe); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	if err := os.Remove(probe); err != nil {
		t.Fatal(err)
	}
	status, output = runFixtureCLI(t, f, "link", "fixtures", "plain")
	if status != 0 || !strings.Contains(output, "Linking:") {
		t.Fatalf("real link failed: %d %q", status, output)
	}
	linked := filepath.Join(f.Target, "README.md")
	info, err := os.Lstat(linked)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("real link missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(f.DataBase, "alma")); err != nil {
		t.Fatalf("app data absent: %v", err)
	}
	status, output = runFixtureCLI(t, f, "link", "fixtures", "plain")
	if status != 0 || !strings.Contains(output, "Item already exists: "+linked) {
		t.Fatalf("repeat conflict: %d %q", status, output)
	}
	record := loadCLIHistory(t, f, "plain")
	if len(record.Destinations) != 1 || record.Destinations[0] != linked {
		t.Fatalf("pre-existing/repeat conflict adopted into history: %+v", record)
	}
	if content, err := os.ReadFile(filepath.Join(f.Target, "keep")); err != nil || string(content) != "pre-existing" {
		t.Fatalf("pre-existing target changed: %q, %v", content, err)
	}
}

func TestLinkEmptyHistoryDryRunAndPersistenceFailure(t *testing.T) {
	f := linkFixture(t)
	f.WriteFile(t, "repository/empty/.alma.yml", "platforms: [{platform: default}]\n")
	status, output := runFixtureCLI(t, f, "link", "fixtures", "empty", "-d")
	if status != 0 || !strings.Contains(output, "Dry run") {
		t.Fatalf("empty dry run: %d %q", status, output)
	}
	if _, err := os.Lstat(filepath.Join(f.DataBase, "alma")); !os.IsNotExist(err) {
		t.Fatalf("dry run wrote history: %v", err)
	}
	status, output = runFixtureCLI(t, f, "link", "fixtures", "empty")
	if status != 0 {
		t.Fatalf("empty real run: %d %q", status, output)
	}
	if got := loadCLIHistory(t, f, "empty"); len(got.Destinations) != 0 {
		t.Fatalf("empty history: %+v", got)
	}

	blocked := linkFixture(t)
	blocked.WriteFile(t, "repository/failure/file", "payload")
	blocked.WriteFile(t, "data/alma/history-v1", "not a directory")
	status, output = runFixtureCLI(t, blocked, "link", "fixtures", "failure")
	if status == 0 || !strings.Contains(output, "save link history") {
		t.Fatalf("persistence failure: %d %q", status, output)
	}
	info, err := os.Lstat(filepath.Join(blocked.Target, "file"))
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("link rolled back after history failure: %v", err)
	}
}

func loadCLIHistory(t *testing.T, f *testutil.Fixture, moduleName string) history.Record {
	t.Helper()
	recordFile, err := history.File(filepath.Join(f.DataBase, "alma"), filepath.Join(f.Repository, moduleName), f.Target)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(recordFile)
	if err != nil {
		t.Fatal(err)
	}
	var record history.Record
	if err := yaml.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	return record
}
