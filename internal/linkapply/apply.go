// Package linkapply creates planned symlinks under caller-selected targets.
package linkapply

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ADIX7/alma2/internal/linkplan"
)

// Apply preserves existing destinations and continues after each filesystem
// failure. Created contains only newly created symlinks for S18 persistence.
// Callers must validate the final plan before calling Apply.
func Apply(candidates []linkplan.Candidate, stdout io.Writer) ([]linkplan.Candidate, error) {
	var created []linkplan.Candidate
	var failures []error
	for _, candidate := range candidates {
		_, err := os.Lstat(candidate.Target)
		if err == nil {
			fmt.Fprintf(stdout, "Item already exists: %s\n", candidate.Target)
			continue
		}
		if !errors.Is(err, os.ErrNotExist) {
			failures = append(failures, report(stdout, candidate, err))
			continue
		}
		sourceInfo, err := os.Lstat(candidate.Source)
		if err != nil {
			failures = append(failures, report(stdout, candidate, err))
			continue
		}
		if sourceInfo.Mode()&os.ModeSymlink != 0 || sourceInfo.IsDir() != candidate.Directory || (!sourceInfo.IsDir() && !sourceInfo.Mode().IsRegular()) {
			failures = append(failures, report(stdout, candidate, fmt.Errorf("source type changed or is a symlink")))
			continue
		}
		if err := os.MkdirAll(filepath.Dir(candidate.Target), 0755); err != nil {
			failures = append(failures, report(stdout, candidate, err))
			continue
		}
		if err := os.Symlink(candidate.Source, candidate.Target); err != nil {
			// A destination created concurrently is still a non-overwriting
			// conflict. Never adopt it into newly created history.
			if _, statErr := os.Lstat(candidate.Target); statErr == nil {
				fmt.Fprintf(stdout, "Item already exists: %s\n", candidate.Target)
				continue
			}
			failures = append(failures, report(stdout, candidate, err))
			continue
		}
		fmt.Fprintf(stdout, "Linking: '%s' '%s'\n", candidate.Source, candidate.Target)
		created = append(created, candidate)
	}
	return created, errors.Join(failures...)
}

func report(stdout io.Writer, candidate linkplan.Candidate, err error) error {
	failure := fmt.Errorf("create link %q -> %q: %w", candidate.Target, candidate.Source, err)
	fmt.Fprintln(stdout, failure)
	if runtime.GOOS == "windows" {
		fmt.Fprintln(stdout, "On Windows symlinks may require Administrator privileges or Developer Mode.")
	}
	return failure
}
