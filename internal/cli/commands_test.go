package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ADIX7/alma2/internal/testutil"
)

func commandFixture(t *testing.T, yaml string) *testutil.Fixture {
	t.Helper()
	f := testutil.NewFixture(t)
	f.WriteFile(t, "config/alma/repository.yml", fmt.Sprintf("repositories: [{name: fixtures, repositoryPath: '%s', linkPath: '%s'}]\n", filepath.ToSlash(f.Repository), filepath.ToSlash(f.Target)))
	f.WriteFile(t, "repository/mod/.alma.yml", yaml)
	return f
}

func TestInstallConfigureExplicitHarmlessCommands(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("native sh integration requires Linux")
	}
	f := commandFixture(t, "platforms: [{platform: default, install: 'printf installed > install-marker', configure: 'printf configured > configure-marker'}]\n")
	f.WriteFile(t, "repository/mod/payload", "do not link")
	status, output := runFixtureCLI(t, f, "configure", "fixtures", "mod")
	if status != 0 || !strings.Contains(output, "printf configured > configure-marker") {
		t.Fatalf("configure: %d %q", status, output)
	}
	if _, err := os.Stat(filepath.Join(f.Work, "configure-marker")); err != nil {
		t.Fatalf("effective cwd not used: %v", err)
	}
	if _, err := os.Stat(filepath.Join(f.Work, "install-marker")); !os.IsNotExist(err) {
		t.Fatalf("install implicitly ran: %v", err)
	}
	status, output = runFixtureCLI(t, f, "install", "fixtures", "mod")
	if status != 0 || !strings.Contains(output, "printf installed > install-marker") {
		t.Fatalf("install: %d %q", status, output)
	}
	if _, err := os.Stat(filepath.Join(f.Work, "install-marker")); err != nil {
		t.Fatalf("install did not run in workdir: %v", err)
	}
	for _, path := range []string{filepath.Join(f.Target, "payload"), filepath.Join(f.DataBase, "alma")} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("command caused implicit link/history: %q, %v", path, err)
		}
	}
}

func TestConfiguredCommandMissingAndUnavailable(t *testing.T) {
	f := commandFixture(t, "platforms: [{platform: default}]\n")
	f.WriteFile(t, "repository/plain/file", "payload")
	f.WriteFile(t, "repository/unsupported/.alma.yml", "platforms: [{platform: never-supported, install: 'printf never'}]\n")
	for _, tc := range []struct {
		args    []string
		message string
		status  int
	}{
		{[]string{"install", "fixtures", "mod"}, "No install command", 0},
		{[]string{"configure", "fixtures", "mod"}, "No configure command", 0},
		{[]string{"install", "fixtures", "plain"}, "No configuration", 0},
		{[]string{"install", "fixtures", "unsupported"}, "unavailable", 1},
		{[]string{"install", "missing", "mod"}, "unknown repository", 1},
		{[]string{"install", "fixtures", "missing"}, "module directory", 1},
	} {
		status, output := runFixtureCLI(t, f, tc.args...)
		if status != tc.status || !strings.Contains(output, tc.message) {
			t.Fatalf("%v: %d %q; want %q", tc.args, status, output, tc.message)
		}
	}
}

func TestConfiguredCommandRejectsMultilineAndPropagatesExit(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("native sh integration requires Linux")
	}
	f := commandFixture(t, `platforms:
  - platform: default
    install: 'exit 7'
    configure: |
      printf first > marker
      echo second
`)
	status, output := runFixtureCLI(t, f, "configure", "fixtures", "mod")
	if status == 0 || !strings.Contains(output, "multiline") {
		t.Fatalf("multiline ran: %d %q", status, output)
	}
	if _, err := os.Stat(filepath.Join(f.Work, "marker")); !os.IsNotExist(err) {
		t.Fatalf("multiline wrote file: %v", err)
	}
	status, output = runFixtureCLI(t, f, "install", "fixtures", "mod")
	if status == 0 || !strings.Contains(output, "exit status 7") {
		t.Fatalf("child failure swallowed: %d %q", status, output)
	}
}
