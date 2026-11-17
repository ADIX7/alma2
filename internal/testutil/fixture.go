// Package testutil provides disposable roots and isolated CLI subprocesses for tests.
package testutil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// Fixture owns all paths that a test CLI invocation should use or modify.
// NewFixture registers removal of Root with testing.TB's cleanup mechanism.
type Fixture struct {
	Root, Home, ConfigBase, DataBase, Repository, Target, Work, Bin, Temp string
}

func NewFixture(t testing.TB) *Fixture {
	t.Helper()
	f := &Fixture{Root: t.TempDir()}
	f.Home = filepath.Join(f.Root, "home")
	f.ConfigBase = filepath.Join(f.Root, "config")
	f.DataBase = filepath.Join(f.Root, "data")
	f.Repository = filepath.Join(f.Root, "repository")
	f.Target = filepath.Join(f.Root, "target")
	f.Work = filepath.Join(f.Root, "work")
	f.Bin = filepath.Join(f.Root, "bin")
	f.Temp = filepath.Join(f.Root, "tmp")
	for _, dir := range []string{f.Home, f.ConfigBase, f.DataBase, f.Repository, f.Target, f.Work, f.Bin, f.Temp} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatalf("create fixture directory %q: %v", dir, err)
		}
	}
	return f
}

// WriteFile writes within Root; relative paths that escape it are rejected.
func (f *Fixture) WriteFile(t testing.TB, relative, contents string) string {
	t.Helper()
	if !filepath.IsLocal(relative) {
		t.Fatalf("unsafe fixture path %q", relative)
	}
	parent := f.Root
	if dir := filepath.Dir(relative); dir != "." {
		for _, part := range strings.Split(dir, string(filepath.Separator)) {
			parent = filepath.Join(parent, part)
			if err := os.Mkdir(parent, 0700); err != nil && !os.IsExist(err) {
				t.Fatalf("create fixture parent %q: %v", parent, err)
			}
			info, err := os.Lstat(parent)
			if err != nil || !info.IsDir() {
				t.Fatalf("fixture parent is not a real directory %q: %v", parent, err)
			}
		}
	}
	path := filepath.Join(f.Root, relative)
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("fixture file is a symlink: %q", path)
	} else if err != nil && !os.IsNotExist(err) {
		t.Fatalf("inspect fixture file %q: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatalf("write fixture file: %v", err)
	}
	return path
}

// Result captures process output and its exit code. A nonzero exit is not a
// runner error; failures to start or wait for the process are.
type Result struct {
	Stdout, Stderr string
	ExitCode       int
}

// RunCLI executes an explicit binary (never through a shell) from the fixture
// working directory. It does not inherit the operator's environment. Keep
// scripts under test harmless and use fake executables in f.Bin, not host tools.
func (f *Fixture) RunCLI(executable string, args ...string) (Result, error) {
	if !filepath.IsAbs(executable) {
		return Result{}, fmt.Errorf("executable must be absolute: %q", executable)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = f.Work
	cmd.Env = []string{
		"HOME=" + f.Home,
		"USERPROFILE=" + f.Home,
		"XDG_CONFIG_HOME=" + f.ConfigBase,
		"XDG_DATA_HOME=" + f.DataBase,
		"ALMA_CONFIG=" + f.ConfigBase,
		"ALMA_CONFIG_FALLBACK=" + f.ConfigBase,
		"ALMA_APP_DATA=" + f.DataBase,
		"ALMA_APP_DATA_FALLBACK=" + f.DataBase,
		"ALMA_WORKDIR=" + f.Work,
		"WORKDIR=" + f.Work,
		"APPDATA=" + f.DataBase,
		"LOCALAPPDATA=" + f.DataBase,
		"TMPDIR=" + f.Temp,
		"TMP=" + f.Temp,
		"TEMP=" + f.Temp,
		"PATH=" + f.Bin,
		"ALMA_TEST_FIXTURE_ROOT=" + f.Root,
	}
	// Windows process creation may need SystemRoot for system DLL lookup.
	if runtime.GOOS == "windows" {
		if systemRoot := os.Getenv("SystemRoot"); systemRoot != "" {
			cmd.Env = append(cmd.Env, "SystemRoot="+systemRoot)
		}
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	result := Result{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: 0}
	if ctx.Err() != nil {
		return result, fmt.Errorf("fixture CLI timed out: %w", ctx.Err())
	}
	if err == nil {
		return result, nil
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		result.ExitCode = exit.ExitCode()
		return result, nil
	}
	return result, fmt.Errorf("run fixture CLI %q: %w", executable, err)
}
