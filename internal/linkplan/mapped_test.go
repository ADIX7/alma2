package linkplan

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/ADIX7/alma2/internal/config"
	"github.com/ADIX7/alma2/internal/pathresolve"
	"github.com/ADIX7/alma2/internal/testutil"
)

func mappedLink(source, target string) config.Link {
	return config.Link{Source: source, Target: target}
}

func TestMappedFilesDirectoriesAndTargets(t *testing.T) {
	f := testutil.NewFixture(t)
	root := filepath.Join(f.Repository, "mod")
	f.WriteFile(t, "repository/mod/editor/config", "editor")
	f.WriteFile(t, "repository/mod/other/config", "other")
	f.WriteFile(t, "repository/mod/whole/child", "whole")
	f.WriteFile(t, "repository/mod/unmapped", "ordinary")
	if err := os.Mkdir(filepath.Join(root, "empty"), 0700); err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(f.Target, "override") // final module target, not repository target
	paths := pathresolve.Resolver{OS: runtime.GOOS, Home: f.Home, Documents: filepath.Join(f.Home, "Documents")}
	links := []config.Link{
		{Source: "editor/config", Target: "renamed/file"},
		{Source: "whole", Target: "whole-linked"},
		{Source: "whole/child", Target: "ignored-beneath-whole"},
		{Source: "empty", Target: "~/empty-linked"},
		{Source: "unmapped", Target: filepath.ToSlash(filepath.Join(f.Target, "absolute-file"))},
	}
	got, err := Mapped(root, base, links, paths)
	want := []Candidate{
		{Source: filepath.Join(root, "editor", "config"), Target: filepath.Join(base, "renamed", "file"), RelativeSource: "editor/config"},
		{Source: filepath.Join(root, "empty"), Target: filepath.Join(f.Home, "empty-linked"), RelativeSource: "empty", Directory: true},
		{Source: filepath.Join(root, "other", "config"), Target: filepath.Join(base, "other", "config"), RelativeSource: "other/config"},
		{Source: filepath.Join(root, "unmapped"), Target: filepath.Join(f.Target, "absolute-file"), RelativeSource: "unmapped"},
		{Source: filepath.Join(root, "whole"), Target: filepath.Join(base, "whole-linked"), RelativeSource: "whole", Directory: true},
	}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("candidates = %+v, %v; want %+v", got, err, want)
	}
	if _, err := os.Stat(base); !os.IsNotExist(err) {
		t.Fatalf("planning created module destination: %v", err)
	}
	if _, err := os.Stat(filepath.Join(f.Home, "empty-linked")); !os.IsNotExist(err) {
		t.Fatalf("planning created home target: %v", err)
	}
}

func TestMappedLastSourceWinsAndYAMLMerge(t *testing.T) {
	f := testutil.NewFixture(t)
	root := filepath.Join(f.Repository, "mod")
	f.WriteFile(t, "repository/mod/.alma.yml", `platforms:
  - platform: default
    links: [{source: file, target: original}]
  - platform: linux
    links: [{source: file, target: renamed}]
`)
	f.WriteFile(t, "repository/mod/file", "payload")
	module, err := config.LoadModule(root)
	if err != nil {
		t.Fatal(err)
	}
	settings := config.Select(module, "linux-x64")
	links := append(settings.Links, config.Link{Source: "./file", Target: "last"})
	got, err := Mapped(root, f.Target, links, pathresolve.Resolver{OS: "linux"})
	if err != nil || len(got) != 2 || got[1].RelativeSource != "file" || got[1].Target != filepath.Join(f.Target, "last") {
		// .alma.yml is still a pre-exclusion candidate until S15.
		t.Fatalf("last mapping not applied: %+v, %v", got, err)
	}
}

func TestMappedCollisionRejectsWholePlan(t *testing.T) {
	f := testutil.NewFixture(t)
	root := filepath.Join(f.Repository, "mod")
	f.WriteFile(t, "repository/mod/a", "a")
	f.WriteFile(t, "repository/mod/b", "b")
	f.WriteFile(t, "repository/mod/folder/file", "nested")
	paths := pathresolve.Resolver{OS: "linux"}
	for _, tc := range []struct {
		name  string
		links []config.Link
	}{
		{"two explicit mappings", []config.Link{mappedLink("a", "same"), mappedLink("b", "same")}},
		{"explicit and automatic", []config.Link{mappedLink("a", "b")}},
		{"cleaned target aliases", []config.Link{mappedLink("a", "folder/../same"), mappedLink("b", "same")}},
		{"mapped directory and file", []config.Link{mappedLink("folder", "same"), mappedLink("b", "same")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			candidates, err := Mapped(root, f.Target, tc.links, paths)
			if err != nil {
				t.Fatalf("mapping: %v", err)
			}
			if err := ValidateDestinations(candidates); err == nil || !strings.Contains(err.Error(), "duplicate destination") {
				t.Fatalf("collision accepted: %+v, %v", candidates, err)
			}
			if _, err := os.Stat(filepath.Join(f.Target, "same")); !os.IsNotExist(err) {
				t.Fatalf("collision created destination: %v", err)
			}
		})
	}
}

func TestValidateDestinationsAfterExclusions(t *testing.T) {
	f := testutil.NewFixture(t)
	candidates := []Candidate{
		{Source: filepath.Join(f.Repository, "a"), Target: filepath.Join(f.Target, "same")},
		{Source: filepath.Join(f.Repository, "b"), Target: filepath.Join(f.Target, "alias", "..", "same")},
	}
	if err := ValidateDestinations(candidates); err == nil {
		t.Fatal("distinct sources sharing a destination accepted")
	}
	// S15 filters exclusions before validating the final plan.
	if err := ValidateDestinations(candidates[:1]); err != nil {
		t.Fatalf("excluded conflicting source still rejected: %v", err)
	}
	f.WriteFile(t, "target/same", "pre-existing target")
	if err := ValidateDestinations(candidates[:1]); err != nil {
		t.Fatalf("single candidate against pre-existing target is not a plan collision: %v", err)
	}
}

func TestMappedRejectsInvalidSourceAndUnsupportedTarget(t *testing.T) {
	f := testutil.NewFixture(t)
	root := filepath.Join(f.Repository, "mod")
	f.WriteFile(t, "repository/mod/file", "content")
	paths := pathresolve.Resolver{OS: "linux"}
	for _, tc := range []struct{ source, target, message string }{
		{"../outside", "target", "cannot escape"},
		{"/absolute", "target", "cannot escape"},
		{"C:/outside", "target", "cannot escape"},
		{`foo\bar`, "target", "must use /"},
		{"", "target", "cannot escape"},
		{"file", "", "target required"},
		{"file", "%APPDATA%/file", "unsupported"},
	} {
		if _, err := Mapped(root, f.Target, []config.Link{{Source: tc.source, Target: tc.target}}, paths); err == nil || !strings.Contains(err.Error(), tc.message) {
			t.Fatalf("mapping %q -> %q: %v; want %q", tc.source, tc.target, err, tc.message)
		}
	}
}

func TestMappedSymlinkSourcesSkipped(t *testing.T) {
	f := testutil.NewFixture(t)
	root := filepath.Join(f.Repository, "mod")
	f.WriteFile(t, "repository/mod/real/file", "payload")
	if err := os.Symlink(filepath.Join(root, "real", "file"), filepath.Join(root, "alias")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	if err := os.Symlink(filepath.Join(root, "real"), filepath.Join(root, "dir-alias")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	got, err := Mapped(root, f.Target, []config.Link{mappedLink("alias", "a"), mappedLink("dir-alias", "b")}, pathresolve.Resolver{OS: "linux"})
	if err != nil || !reflect.DeepEqual(got, []Candidate{{Source: filepath.Join(root, "real", "file"), Target: filepath.Join(f.Target, "real", "file"), RelativeSource: "real/file"}}) {
		t.Fatalf("symlink mapped: %+v, %v", got, err)
	}
}
