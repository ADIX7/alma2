package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFixtureCleanupAndFiles(t *testing.T) {
	var root string
	t.Run("disposable", func(t *testing.T) {
		f := NewFixture(t)
		root = f.Root
		for _, dir := range []string{f.Home, f.ConfigBase, f.DataBase, f.Repository, f.Target, f.Work, f.Bin, f.Temp} {
			if info, err := os.Stat(dir); err != nil || !info.IsDir() {
				t.Fatalf("fixture directory %q: %v", dir, err)
			}
		}
		path := f.WriteFile(t, "repository/module/.alma.yml", "platforms: []\n")
		if data, err := os.ReadFile(path); err != nil || string(data) != "platforms: []\n" {
			t.Fatalf("fixture file: %q, %v", data, err)
		}
	})
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("fixture root %q survived cleanup: %v", root, err)
	}
	if filepath.IsLocal("../outside") || filepath.IsLocal("/outside") {
		t.Fatal("unsafe fixture paths unexpectedly local")
	}
}

// Executed only by the controlled subprocess test below, not by normal go test.
func TestFixtureProbeProcess(t *testing.T) {
	root := os.Getenv("ALMA_TEST_FIXTURE_ROOT")
	if root == "" {
		return
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if cwd != filepath.Join(root, "work") || os.Getenv("ALMA_TEST_SECRET") != "" {
		t.Fatalf("inherited environment or cwd: cwd=%q secret=%q", cwd, os.Getenv("ALMA_TEST_SECRET"))
	}
	for name, dir := range map[string]string{
		"HOME":            filepath.Join(root, "home"),
		"USERPROFILE":     filepath.Join(root, "home"),
		"ALMA_CONFIG":     filepath.Join(root, "config"),
		"XDG_CONFIG_HOME": filepath.Join(root, "config"),
		"ALMA_APP_DATA":   filepath.Join(root, "data"),
		"LOCALAPPDATA":    filepath.Join(root, "data"),
		"ALMA_WORKDIR":    filepath.Join(root, "work"),
		"TMPDIR":          filepath.Join(root, "tmp"),
	} {
		if got := os.Getenv(name); got != dir {
			t.Fatalf("%s=%q, want %q", name, got, dir)
		}
	}
	for _, dir := range []string{os.Getenv("HOME"), os.Getenv("ALMA_CONFIG"), os.Getenv("ALMA_APP_DATA")} {
		if err := os.WriteFile(filepath.Join(dir, "probe"), []byte("safe"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	fmt.Fprintln(os.Stdout, "isolated stdout")
	fmt.Fprintln(os.Stderr, "isolated stderr")
	os.Exit(7)
}

func TestRunCLIIsolation(t *testing.T) {
	f := NewFixture(t)
	t.Setenv("ALMA_TEST_SECRET", "never inherit")
	operatorHome := t.TempDir()
	operatorConfig := t.TempDir()
	operatorData := t.TempDir()
	t.Setenv("HOME", operatorHome)
	t.Setenv("ALMA_CONFIG", operatorConfig)
	t.Setenv("ALMA_APP_DATA", operatorData)
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	result, err := f.RunCLI(binary, "-test.run=^TestFixtureProbeProcess$")
	if err != nil {
		t.Fatal(err)
	}
	if result.ExitCode != 7 || !strings.Contains(result.Stdout, "isolated stdout") || !strings.Contains(result.Stderr, "isolated stderr") {
		t.Fatalf("unexpected result: %+v", result)
	}
	for _, dir := range []string{f.Home, f.ConfigBase, f.DataBase} {
		if data, err := os.ReadFile(filepath.Join(dir, "probe")); err != nil || string(data) != "safe" {
			t.Fatalf("probe file in %q: %q, %v", dir, data, err)
		}
	}
	for _, dir := range []string{operatorHome, operatorConfig, operatorData} {
		if _, err := os.Stat(filepath.Join(dir, "probe")); !os.IsNotExist(err) {
			t.Fatalf("subprocess wrote outside fixture %q: %v", dir, err)
		}
	}
	if _, err := f.RunCLI("relative-executable"); err == nil {
		t.Fatal("relative executable accepted")
	}
	if _, err := f.RunCLI(filepath.Join(f.Bin, "missing")); err == nil {
		t.Fatal("missing executable accepted")
	}
}
