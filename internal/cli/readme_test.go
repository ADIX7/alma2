package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ADIX7/alma2/internal/testutil"
)

func TestReadmePriorityAndLiteralOutput(t *testing.T) {
	f := testutil.NewFixture(t)
	f.WriteFile(t, "config/alma/repository.yml", fmt.Sprintf("repositories: [{name: fixtures, repositoryPath: '%s'}]\n", filepath.ToSlash(f.Repository)))
	f.WriteFile(t, "repository/mod/.alma.yml", "platforms: [{platform: never-supported}]\n")
	f.WriteFile(t, "repository/mod/README", "plain\n")
	f.WriteFile(t, "repository/mod/README.txt", "text\n")
	f.WriteFile(t, "repository/mod/README.md", "# literal **Markdown**")
	for _, tc := range []struct{ removed, want string }{
		{"", "# literal **Markdown**"},
		{"README.md", "text\n"},
		{"README.txt", "plain\n"},
	} {
		if tc.removed != "" {
			if err := os.Remove(filepath.Join(f.Repository, "mod", tc.removed)); err != nil {
				t.Fatal(err)
			}
		}
		status, output := runFixtureCLI(t, f, "readme", "fixtures", "mod")
		if status != 0 || output != tc.want {
			t.Fatalf("priority %q: %d %q; want %q", tc.removed, status, output, tc.want)
		}
	}
}

func TestReadmeNoConfigCaseAndErrors(t *testing.T) {
	f := testutil.NewFixture(t)
	f.WriteFile(t, "work/plain/README.txt", "plain fixture\n")
	status, output := runFixtureCLI(t, f, "readme", "plain")
	if status != 0 || output != "plain fixture\n" {
		t.Fatalf("no config readme: %d %q", status, output)
	}
	f.WriteFile(t, "work/other/readme.md", "wrong case")
	f.WriteFile(t, "work/other/nested/README.md", "nested should not count")
	status, output = runFixtureCLI(t, f, "readme", "other")
	if status != 0 || !strings.Contains(output, "No README file found") {
		t.Fatalf("missing root README: %d %q", status, output)
	}
	status, output = runFixtureCLI(t, f, "readme", "missing")
	if status == 0 || !strings.Contains(output, "module directory") {
		t.Fatalf("missing module: %d %q", status, output)
	}
	status, output = runFixtureCLI(t, f, "readme", "../plain")
	if status == 0 || !strings.Contains(output, "unsafe module") {
		t.Fatalf("unsafe module: %d %q", status, output)
	}
}
