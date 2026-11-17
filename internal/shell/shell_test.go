package shell

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/ADIX7/alma2/internal/testutil"
)

type fakeRunner struct {
	calls     int
	name, cwd string
	args      []string
	failure   error
}

func (f *fakeRunner) Run(_ context.Context, name string, args []string, cwd string, stdout, stderr io.Writer) error {
	f.calls++
	f.name, f.args, f.cwd = name, args, cwd
	fmt.Fprintln(stdout, "line")
	fmt.Fprintln(stderr, "diagnostic")
	return f.failure
}

func TestShellSelectionAndValidation(t *testing.T) {
	f := testutil.NewFixture(t)
	for _, tc := range []struct{ os, name string }{{"linux", "sh"}, {"windows", "pwsh"}} {
		runner := &fakeRunner{}
		var out, stderr bytes.Buffer
		if err := Execute(context.Background(), tc.os, "printf hello", f.Work, runner, &out, &stderr); err != nil {
			t.Fatal(err)
		}
		if runner.calls != 1 || runner.name != tc.name || runner.cwd != f.Work || len(runner.args) != 2 || runner.args[0] != "-c" || runner.args[1] != "printf hello" || out.String() != "line\n" || stderr.String() != "diagnostic\n" {
			t.Fatalf("%s: %+v, %q, %q", tc.os, runner, out.String(), stderr.String())
		}
	}
	for _, tc := range []struct{ os, command, message string }{
		{"darwin", "echo harmless", "unsupported"},
		{"linux", "echo a\necho b", "multiline"},
		{"windows", "echo a\r\necho b", "multiline"},
		{"windows", "echo a\necho b", "multiline"},
	} {
		runner := &fakeRunner{}
		if err := Execute(context.Background(), tc.os, tc.command, f.Work, runner, io.Discard, io.Discard); err == nil || !strings.Contains(err.Error(), tc.message) || runner.calls != 0 {
			t.Fatalf("%s: %v, calls=%d", tc.os, err, runner.calls)
		}
	}
	failure := errors.New("start failed")
	runner := &fakeRunner{failure: failure}
	if err := Execute(context.Background(), "linux", "echo harmless", f.Work, runner, io.Discard, io.Discard); !errors.Is(err, failure) {
		t.Fatalf("start failure lost: %v", err)
	}
}

func TestProcessRunnerHarmlessFixtureCommands(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("native sh smoke test requires Linux")
	}
	f := testutil.NewFixture(t)
	var out, stderr bytes.Buffer
	if err := Execute(context.Background(), "linux", "printf 'hello\\n'; pwd; printf 'diagnostic\\n' >&2", f.Work, ProcessRunner{}, &out, &stderr); err != nil {
		t.Fatal(err)
	}
	if out.String() != "hello\n"+f.Work+"\n" || stderr.String() != "diagnostic\n" {
		t.Fatalf("stdout/stderr/cwd = %q, %q", out.String(), stderr.String())
	}
	out.Reset()
	err := Execute(context.Background(), "linux", "exit 7", f.Work, ProcessRunner{}, &out, &stderr)
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 7 {
		t.Fatalf("exit code not propagated: %v", err)
	}
	out.Reset()
	if err := Execute(context.Background(), "linux", "read x; printf 'eof\\n'", f.Work, ProcessRunner{}, &out, &stderr); err != nil || out.String() != "eof\n" {
		t.Fatalf("stdin was not redirected to EOF: %v, %q", err, out.String())
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Execute(ctx, "linux", "printf 'not run'", f.Work, ProcessRunner{}, &out, &stderr); err == nil {
		t.Fatal("cancelled process unexpectedly succeeded")
	}
}
