// Package cli handles Alma command selection and positional arguments.
package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

var commands = map[string]bool{
	"configure": true, "diag": true, "help": true, "info": true,
	"install": true, "link": true, "ls": true, "readme": true, "unlink": true,
}

// Invocation contains the parsed arguments needed by future command handlers.
type Invocation struct {
	Command, Repository, Module, Helper string
	DryRun, LinkHelp                    bool
}

// Parse selects a case-sensitive command. Only module commands and info discard
// tokens beginning with '-'; ls and diag retain their positional tokens.
func Parse(args []string) (Invocation, error) {
	if len(args) == 0 {
		return Invocation{}, fmt.Errorf("No command was given")
	}
	name := args[0]
	if name == "--help" || name == "-h" {
		name = "help"
	}
	if !commands[name] {
		return Invocation{}, fmt.Errorf("Invalid command: %s", args[0])
	}
	inv := Invocation{Command: name}
	positional := args[1:]
	switch name {
	case "link", "install", "configure", "readme", "info":
		positional = make([]string, 0, len(args)-1)
		for _, arg := range args[1:] {
			if strings.HasPrefix(arg, "-") {
				if name == "link" {
					inv.DryRun = inv.DryRun || arg == "-d" || arg == "--dry-run"
					inv.LinkHelp = inv.LinkHelp || arg == "--help"
				}
				continue
			}
			positional = append(positional, arg)
		}
	}
	switch name {
	case "link", "install", "configure", "readme":
		if len(positional) == 0 && !inv.LinkHelp {
			return Invocation{}, fmt.Errorf("No module specified")
		}
		if len(positional) >= 2 {
			inv.Repository, inv.Module = positional[0], positional[1]
		} else if len(positional) == 1 {
			inv.Module = positional[0]
		}
	case "info":
		if len(positional) > 0 {
			inv.Repository = positional[0]
		}
		if len(positional) > 1 {
			inv.Module = positional[1]
		}
	case "ls":
		if len(positional) > 0 {
			inv.Repository = positional[0]
		}
	case "diag":
		if len(positional) > 0 {
			inv.Helper = positional[0]
		}
	}
	return inv, nil
}

// Run writes ordinary messages to stdout and returns a process exit code.
// Commands not yet implemented fail explicitly; no placeholder performs work.
func Run(args []string, stdout io.Writer) int {
	inv, err := Parse(args)
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}
	if inv.Command == "link" {
		return runLink(inv, stdout)
	}
	if inv.Command == "ls" {
		return runList(inv, stdout)
	}
	if inv.Command == "help" {
		names := make([]string, 0, len(commands))
		for name := range commands {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Fprintln(stdout, name)
		}
		return 0
	}
	fmt.Fprintf(stdout, "Command not implemented: %s\n", inv.Command)
	return 1
}
