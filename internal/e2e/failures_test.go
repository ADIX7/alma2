package e2e

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ADIX7/alma2/internal/history"
	"github.com/ADIX7/alma2/internal/testutil"
	"gopkg.in/yaml.v3"
)

func rejected(t *testing.T, f *testutil.Fixture, binary, message string, args ...string) {
	t.Helper()
	result, err := f.RunCLI(binary, args...)
	if err != nil || result.ExitCode == 0 || !strings.Contains(result.Stdout, message) {
		t.Fatalf("alma %v: %+v, %v; want %q", args, result, err, message)
	}
}

func TestSubprocessRejectedConfigurationAndPlanAreReadOnly(t *testing.T) {
	f := testutil.NewFixture(t)
	binary := buildFixtureCLI(t, f)
	f.WriteFile(t, "config/alma/repository.json", `{ "repositories": [{ "name": "legacy" }] }`)
	if output := invoke(t, f, binary, "ls"); output != "Repositories\n" {
		t.Fatalf("JSON fallback: %q", output)
	}
	rejected(t, f, binary, "unknown repository", "ls", "legacy")
	f.WriteFile(t, "config/alma/repository.yml", "repositories: null\n")
	rejected(t, f, binary, "null", "ls")
	fixtureFile(t, f, "repository.yml", "config/alma/repository.yml")
	f.WriteFile(t, "repository/mod/.alma-config.json", `{ "platforms": [] }`)
	f.WriteFile(t, "repository/mod/file", "payload")
	jsonOnly := invoke(t, f, binary, "link", "fixtures", "mod", "-d")
	if !strings.Contains(jsonOnly, filepath.Join(f.Repository, "mod", "file")) || strings.Contains(jsonOnly, filepath.Join(f.Repository, "mod", ".alma-config.json")) {
		t.Fatalf("JSON-only module interpreted as config/payload: %q", jsonOnly)
	}
	rejected(t, f, binary, "unknown repository", "link", "unknown", "mod", "-d")
	rejected(t, f, binary, "unsafe module", "link", "fixtures", "../mod", "-d")
	rejected(t, f, binary, "unsafe module", "link", "fixtures", "mod/nested", "-d")
	f.WriteFile(t, "repository/mod/.alma.yml", "platforms: [{platform: default, target: ~}]\n")
	rejected(t, f, binary, "null", "link", "fixtures", "mod", "-d")
	f.WriteFile(t, "repository/mod/.alma.yml", "platforms: [{platform: never-supported}]\n")
	rejected(t, f, binary, "unavailable", "link", "fixtures", "mod", "-d")
	f.WriteFile(t, "repository/mod/.alma.yml", "platforms: [{platform: default, links: [{source: file, target: collision}, {source: other, target: collision}]}]\n")
	f.WriteFile(t, "repository/mod/other", "other")
	rejected(t, f, binary, "duplicate destination", "link", "fixtures", "mod", "-d")
	rejected(t, f, binary, "duplicate destination", "link", "fixtures", "mod")
	for _, path := range []string{filepath.Join(f.DataBase, "alma"), filepath.Join(f.Target, "collision")} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("rejected plan wrote %q: %v", path, err)
		}
	}
}

func TestSubprocessPartialFailurePersistsOnlySuccess(t *testing.T) {
	f := testutil.NewFixture(t)
	binary := buildFixtureCLI(t, f)
	fixtureFile(t, f, "repository.yml", "config/alma/repository.yml")
	f.WriteFile(t, "repository/mod/.alma.yml", "platforms: [{platform: default, links: [{source: a, target: blocker/fail}, {source: b, target: good}]}]\n")
	f.WriteFile(t, "repository/mod/a", "a")
	f.WriteFile(t, "repository/mod/b", "b")
	blocker := f.WriteFile(t, "target/blocker", "preserve")
	old := f.WriteFile(t, "data/alma/moduleHashes.txt", "legacy untouched")
	probe := filepath.Join(f.Target, "probe")
	if err := os.Symlink(blocker, probe); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	if err := os.Remove(probe); err != nil {
		t.Fatal(err)
	}
	rejected(t, f, binary, "create link", "link", "fixtures", "mod")
	info, err := os.Lstat(filepath.Join(f.Target, "good"))
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("partial success missing: %v", err)
	}
	if data, err := os.ReadFile(blocker); err != nil || string(data) != "preserve" {
		t.Fatalf("blocker changed: %q, %v", data, err)
	}
	if data, err := os.ReadFile(old); err != nil || string(data) != "legacy untouched" {
		t.Fatalf("old history changed: %q, %v", data, err)
	}
	recordFile, err := history.File(filepath.Join(f.DataBase, "alma"), filepath.Join(f.Repository, "mod"), f.Target)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(recordFile)
	if err != nil {
		t.Fatal(err)
	}
	var record history.Record
	if err := yaml.Unmarshal(data, &record); err != nil || len(record.Destinations) != 1 || record.Destinations[0] != filepath.Join(f.Target, "good") {
		t.Fatalf("partial history: %+v, %v", record, err)
	}
}

func TestSubprocessCommandRestrictionsAndUnlink(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("fake sh subprocess test requires Linux")
	}
	f := testutil.NewFixture(t)
	binary := buildFixtureCLI(t, f)
	fixtureFile(t, f, "repository.yml", "config/alma/repository.yml")
	f.WriteFile(t, "repository/mod/.alma.yml", `platforms:
  - platform: default
    install: 'printf install'
    configure: |
      printf first
      printf second
`)
	f.WriteFile(t, "repository/mod/payload", "never link")
	fake := f.WriteFile(t, "bin/sh", `#!/bin/sh
printf '%s' "$*" > "$ALMA_TEST_FIXTURE_ROOT/work/shell-call"
`)
	if err := os.Chmod(fake, 0700); err != nil {
		t.Fatal(err)
	}
	rejected(t, f, binary, "multiline", "configure", "fixtures", "mod")
	if _, err := os.Lstat(filepath.Join(f.Work, "shell-call")); !os.IsNotExist(err) {
		t.Fatalf("multiline spawned shell: %v", err)
	}
	invoke(t, f, binary, "install", "fixtures", "mod")
	if data, err := os.ReadFile(filepath.Join(f.Work, "shell-call")); err != nil || string(data) != "-c printf install" {
		t.Fatalf("wrong shell invocation: %q, %v", data, err)
	}
	rejected(t, f, binary, "unlink is unsupported", "unlink", "fixtures", "mod")
	for _, path := range []string{filepath.Join(f.Target, "payload"), filepath.Join(f.DataBase, "alma")} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("command/unlink wrote %q: %v", path, err)
		}
	}
}
