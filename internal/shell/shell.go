// Package shell runs explicitly requested, trusted single-line commands.
package shell

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Runner allows shell process creation to be replaced in tests.
type Runner interface {
	Run(ctx context.Context, name string, args []string, cwd string, stdout, stderr io.Writer) error
}

type ProcessRunner struct{}

func (ProcessRunner) Run(ctx context.Context, name string, args []string, cwd string, stdout, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = cwd
	cmd.Stdin = bytes.NewReader(nil) // redirected EOF, never an interactive terminal
	cmd.Stdout, cmd.Stderr = stdout, stderr
	return cmd.Run()
}

// Execute validates a trusted command before selecting the supported shell.
// A child's nonzero exit is returned as an error; no other commands are run.
func Execute(ctx context.Context, osName, command, cwd string, runner Runner, stdout, stderr io.Writer) error {
	if strings.ContainsAny(command, "\r\n") {
		return fmt.Errorf("multiline shell commands are unsupported")
	}
	name := ""
	switch osName {
	case "linux":
		name = "sh"
	case "windows":
		name = "pwsh"
	default:
		return fmt.Errorf("shell execution unsupported on %s", osName)
	}
	if runner == nil {
		runner = ProcessRunner{}
	}
	if err := runner.Run(ctx, name, []string{"-c", command}, cwd, stdout, stderr); err != nil {
		return fmt.Errorf("shell command failed: %w", err)
	}
	return nil
}
