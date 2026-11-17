package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ADIX7/alma2/internal/testutil"
)

func TestDiagnosticsHelpersAndProviderValues(t *testing.T) {
	f := testutil.NewFixture(t)
	status, output := runFixtureCLI(t, f, "diag", "list", "ignored")
	want := "Available diagnostic helpers:\napp-data-providers\nconfig-home-providers\nspecial-path-resolver\n"
	if status != 0 || output != want {
		t.Fatalf("helper order: %d %q", status, output)
	}
	status, output = runFixtureCLI(t, f, "diag", "--help")
	if status != 0 || !strings.Contains(output, "Usage:") {
		t.Fatalf("diag help: %d %q", status, output)
	}
	t.Setenv("ALMA_CONFIG_FALLBACK", "")
	status, output = runFixtureCLI(t, f, "diag", "config-home-providers")
	if status != 0 || !strings.Contains(output, "ALMA_CONFIG => "+f.ConfigBase) || !strings.Contains(output, "ALMA_CONFIG_FALLBACK => <null>") || !strings.Contains(output, "There are 4 providers") {
		t.Fatalf("config providers: %d %q", status, output)
	}
	status, output = runFixtureCLI(t, f, "diag", "app-data-providers")
	if status != 0 || !strings.Contains(output, "ALMA_APP_DATA => "+f.DataBase) || !strings.Contains(output, "There are 3 providers") {
		t.Fatalf("app data providers: %d %q", status, output)
	}
	status, output = runFixtureCLI(t, f, "diag", "special-path-resolver")
	if status != 0 || !strings.Contains(output, "~ => "+f.Home) || !strings.Contains(output, "%DOCUMENTS% => "+filepath.Join(f.Home, "Documents")) || !strings.Contains(output, "%APPDATA% => ") {
		t.Fatalf("special paths: %d %q", status, output)
	}
	if _, err := os.Lstat(filepath.Join(f.DataBase, "alma")); !os.IsNotExist(err) {
		t.Fatalf("diagnostics wrote app data: %v", err)
	}
}

func TestDiagnosticsErrorsAndUnlinkNeverDeletes(t *testing.T) {
	f := testutil.NewFixture(t)
	saved := f.WriteFile(t, "target/keep", "preserve")
	for _, tc := range []struct {
		args    []string
		message string
	}{
		{[]string{"diag"}, "No diagnostic helper"},
		{[]string{"diag", "unknown"}, "is not found"},
		{[]string{"unlink", "keep"}, "unlink is unsupported"},
	} {
		status, output := runFixtureCLI(t, f, tc.args...)
		if status == 0 || !strings.Contains(output, tc.message) {
			t.Fatalf("%v: %d %q", tc.args, status, output)
		}
	}
	contents, err := os.ReadFile(saved)
	if err != nil || string(contents) != "preserve" {
		t.Fatalf("unlink removed/changed file: %q, %v", contents, err)
	}
}
