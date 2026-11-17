package cli

import (
	"bytes"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want Invocation
		err  string
	}{
		{"no command", nil, Invocation{}, "No command was given"},
		{"case sensitive", []string{"Link", "module"}, Invocation{}, "Invalid command: Link"},
		{"unknown", []string{"missing"}, Invocation{}, "Invalid command: missing"},
		{"help alias", []string{"-h", "ignored"}, Invocation{Command: "help"}, ""},
		{"long help alias", []string{"--help"}, Invocation{Command: "help"}, ""},
		{"module required", []string{"configure", "-x"}, Invocation{}, "No module specified"},
		{"link help special", []string{"link", "--help"}, Invocation{Command: "link", LinkHelp: true}, ""},
		{"link short help not special", []string{"link", "-h"}, Invocation{}, "No module specified"},
		{"link dry run not a module", []string{"link", "--dry-run"}, Invocation{}, "No module specified"},
		{"single module", []string{"readme", "fish", "-x"}, Invocation{Command: "readme", Module: "fish"}, ""},
		{"readme extras", []string{"readme", "repo", "fish", "ignored"}, Invocation{Command: "readme", Repository: "repo", Module: "fish"}, ""},
		{"two positions", []string{"install", "-x", "dotfiles", "shell/fish", "extra"}, Invocation{Command: "install", Repository: "dotfiles", Module: "shell/fish"}, ""},
		{"one position", []string{"configure", "fish"}, Invocation{Command: "configure", Module: "fish"}, ""},
		{"link flags", []string{"link", "repo", "mod", "-d", "--dry-run", "ignored"}, Invocation{Command: "link", Repository: "repo", Module: "mod", DryRun: true}, ""},
		{"info general", []string{"info", "-x"}, Invocation{Command: "info"}, ""},
		{"info two positions", []string{"info", "-h", "repo", "mod", "ignored"}, Invocation{Command: "info", Repository: "repo", Module: "mod"}, ""},
		{"ls preserves flags", []string{"ls", "-h", "ignored"}, Invocation{Command: "ls", Repository: "-h"}, ""},
		{"diag helper", []string{"diag", "--help", "ignored"}, Invocation{Command: "diag", Helper: "--help"}, ""},
		{"unlink known", []string{"unlink"}, Invocation{Command: "unlink"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args)
			if tt.err != "" {
				if err == nil || err.Error() != tt.err {
					t.Fatalf("Parse error = %v, want %q", err, tt.err)
				}
				return
			}
			if err != nil || !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Parse = %+v, %v; want %+v", got, err, tt.want)
			}
		})
	}
}

func TestRun(t *testing.T) {
	const help = "configure\ndiag\nhelp\ninfo\ninstall\nlink\nls\nreadme\nunlink\n"
	for _, tt := range []struct {
		name   string
		args   []string
		output string
		status int
	}{
		{"help", []string{"help", "ignored"}, help, 0},
		{"short help", []string{"-h"}, help, 0},
		{"long help", []string{"--help"}, help, 0},
		{"no command", nil, "No command was given\n", 1},
		{"unknown", []string{"HELP"}, "Invalid command: HELP\n", 1},
		{"missing module", []string{"link", "-d"}, "No module specified\n", 1},
		{"link help", []string{"link", "--help"}, "Usage: alma link [repository] <module> [-d|--dry-run]\n", 0},
		{"unimplemented module info", []string{"info", "repo", "mod"}, "Command not implemented: module info\n", 1},
		{"unimplemented unlink", []string{"unlink"}, "Command not implemented: unlink\n", 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			status := Run(tt.args, &output)
			if status != tt.status || output.String() != tt.output {
				t.Fatalf("Run = %d, %q; want %d, %q", status, output.String(), tt.status, tt.output)
			}
		})
	}
}
