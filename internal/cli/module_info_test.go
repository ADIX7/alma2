package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ADIX7/alma2/internal/testutil"
)

func TestModuleInfoResolvedTargetsAndExistence(t *testing.T) {
	f := testutil.NewFixture(t)
	f.WriteFile(t, "config/alma/repository.yml", fmt.Sprintf("repositories: [{name: fixtures, repositoryPath: '%s', linkPath: '%s'}]\n", filepath.ToSlash(f.Repository), filepath.ToSlash(f.Target)))
	f.WriteFile(t, "repository/mod/.alma.yml", `platforms:
  - platform: default
    target: inner
    links:
      - {source: present, target: mapped/existing}
      - {source: missing, target: absent}
    install: 'echo harmless'
    configure: 'echo harmless'
`)
	f.WriteFile(t, "repository/mod/present", "source")
	existing := f.WriteFile(t, "target/inner/mapped/existing", "ordinary file, not Alma link")
	status, output := runFixtureCLI(t, f, "info", "fixtures", "mod")
	if status != 0 || !strings.Contains(output, "Target directory is: "+filepath.Join(f.Target, "inner")) || !strings.Contains(output, "Can be installed") || !strings.Contains(output, "Can be configured") || !strings.Contains(output, "Has 2 link(s)") {
		t.Fatalf("module info: %d %q", status, output)
	}
	if !strings.Contains(output, green+"present"+reset+" -> "+green+existing+reset) || !strings.Contains(output, red+"missing"+reset+" -> "+red+filepath.Join(f.Target, "inner", "absent")+reset) {
		t.Fatalf("actual mapped destination existence/color incorrect: %q", output)
	}
}

func TestModuleInfoNoConfigUnavailableAndErrors(t *testing.T) {
	f := testutil.NewFixture(t)
	f.WriteFile(t, "config/alma/repository.yml", fmt.Sprintf("repositories: [{name: fixtures, repositoryPath: '%s'}]\n", filepath.ToSlash(f.Repository)))
	f.WriteFile(t, "repository/plain/file", "payload")
	f.WriteFile(t, "repository/unsupported/.alma.yml", "platforms: [{platform: never-supported}]\n")
	f.WriteFile(t, "repository/empty/.alma.yml", "platforms: [{platform: default}]\n")
	for _, tc := range []struct {
		args       []string
		message    string
		wantStatus int
	}{
		{[]string{"info", "fixtures", "plain"}, "No configuration is found", 0},
		{[]string{"info", "fixtures", "empty"}, "Has no links.", 0},
		{[]string{"info", "fixtures", "unsupported"}, "Unavailable on platform", 0},
		{[]string{"info", "fixtures", "missing"}, "module directory", 1},
		{[]string{"info", "unknown", "plain"}, "unknown repository", 1},
		{[]string{"info", "fixtures", "../plain"}, "unsafe module", 1},
	} {
		status, output := runFixtureCLI(t, f, tc.args...)
		if status != tc.wantStatus || !strings.Contains(output, tc.message) {
			t.Fatalf("%v: %d %q; want %q", tc.args, status, output, tc.message)
		}
	}
}
