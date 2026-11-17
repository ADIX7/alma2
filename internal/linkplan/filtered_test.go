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

func relativeCandidates(candidates []Candidate) []string {
	names := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		names = append(names, candidate.RelativeSource)
	}
	return names
}

func TestPlanConfiguredAndUnconfiguredPayloads(t *testing.T) {
	f := testutil.NewFixture(t)
	root := filepath.Join(f.Repository, "fish")
	names := []string{
		".alma.yml", ".alma-config.json", ".alma.yaml", "README.md", "README.txt", "README",
		"nested/README.md", "nested/.alma.yml", "nested/.alma-config.json", "nested/file",
	}
	if runtime.GOOS != "windows" { // Windows treats readme.md and README.md as the same file.
		names = append(names, "readme.md")
	}
	for _, name := range names {
		f.WriteFile(t, "repository/fish/"+name, "fixture")
	}
	paths := pathresolve.Resolver{OS: "linux"}
	unconfigured := config.Select(nil, "linux-x64")
	got, err := Plan(root, f.Target, unconfigured, paths)
	want := []string{".alma.yaml", "README", "README.md", "README.txt", "nested/README.md", "nested/file"}
	if runtime.GOOS != "windows" {
		want = append(want, "readme.md")
	}
	if err != nil || !reflect.DeepEqual(relativeCandidates(got), want) {
		t.Fatalf("unconfigured: %v, %v; want %v", relativeCandidates(got), err, want)
	}
	configured := config.Select(&config.Module{Platforms: []config.Platform{{Selector: "default"}}}, "linux-x64")
	got, err = Plan(root, f.Target, configured, paths)
	want = []string{".alma.yaml", "nested/README.md", "nested/file"}
	if runtime.GOOS != "windows" {
		want = append(want, "readme.md")
	}
	if err != nil || !reflect.DeepEqual(relativeCandidates(got), want) {
		t.Fatalf("configured default README: %v, %v; want %v", relativeCandidates(got), err, want)
	}
	if _, err := os.Stat(filepath.Join(f.Target, "nested")); !os.IsNotExist(err) {
		t.Fatalf("planning wrote nested target: %v", err)
	}
}

func TestPlanMergedExclusionsAndMappedDirectoryCaveat(t *testing.T) {
	f := testutil.NewFixture(t)
	root := filepath.Join(f.Repository, "mod")
	f.WriteFile(t, "repository/mod/.alma.yml", `platforms:
  - platform: default
    exclude: [tmp, whole/child]
    links:
      - {source: whole, target: whole-linked}
      - {source: .alma.yml, target: config-copy}
  - platform: linux
    exclude: [sub/secret]
    excludeReadme: false
    links:
      - {source: .alma-config.json, target: legacy-copy}
`)
	for _, name := range []string{".alma-config.json", "README", "tmp/a", "tmp2/a", "sub/secret", "sub/keep", "whole/child", "whole/other"} {
		f.WriteFile(t, "repository/mod/"+name, "fixture")
	}
	module, err := config.LoadModule(root)
	if err != nil {
		t.Fatal(err)
	}
	settings := config.Select(module, "linux-x64")
	got, err := Plan(root, f.Target, settings, pathresolve.Resolver{OS: "linux"})
	want := []string{"README", "sub/keep", "tmp2/a", "whole"}
	if err != nil || !reflect.DeepEqual(relativeCandidates(got), want) {
		t.Fatalf("merged exclusions: %v, %v; want %v", relativeCandidates(got), err, want)
	}
	if !got[3].Directory || got[3].Target != filepath.Join(f.Target, "whole-linked") {
		t.Fatalf("child exclusion must not expand already mapped directory: %+v", got[3])
	}
	settings.Exclude = append(settings.Exclude, "whole")
	got, err = Plan(root, f.Target, settings, pathresolve.Resolver{OS: "linux"})
	if err != nil || !reflect.DeepEqual(relativeCandidates(got), []string{"README", "sub/keep", "tmp2/a"}) {
		t.Fatalf("exclude mapped directory itself: %v, %v", relativeCandidates(got), err)
	}
}

func TestPlanReadmeDirectoryIsNotAReadmeFile(t *testing.T) {
	f := testutil.NewFixture(t)
	root := filepath.Join(f.Repository, "mod")
	f.WriteFile(t, "repository/mod/README.md/payload", "fixture")
	settings := config.Settings{Available: true, ExcludeReadme: true,
		Links: []config.Link{{Source: "README.md", Target: "mapped-readme-folder"}}}
	got, err := Plan(root, f.Target, settings, pathresolve.Resolver{OS: "linux"})
	if err != nil || len(got) != 1 || !got[0].Directory || got[0].RelativeSource != "README.md" {
		t.Fatalf("README-named directory: %+v, %v", got, err)
	}
}

func TestPlanCollisionOnlyAfterExclusions(t *testing.T) {
	f := testutil.NewFixture(t)
	root := filepath.Join(f.Repository, "mod")
	f.WriteFile(t, "repository/mod/a", "a")
	f.WriteFile(t, "repository/mod/b", "b")
	paths := pathresolve.Resolver{OS: "linux"}
	settings := config.Settings{Available: true, Links: []config.Link{{Source: "a", Target: "same"}, {Source: "b", Target: "same"}}}
	if candidates, err := Plan(root, f.Target, settings, paths); err == nil || candidates != nil || !strings.Contains(err.Error(), "duplicate destination") {
		t.Fatalf("unfiltered collision: %+v, %v", candidates, err)
	}
	settings.Exclude = []string{"b"}
	got, err := Plan(root, f.Target, settings, paths)
	if err != nil || !reflect.DeepEqual(relativeCandidates(got), []string{"a"}) {
		t.Fatalf("excluded collision: %v, %v", relativeCandidates(got), err)
	}
	if _, err := os.Stat(filepath.Join(f.Target, "same")); !os.IsNotExist(err) {
		t.Fatalf("plan created collision destination: %v", err)
	}
}

func TestPlanUnavailableAndInvalidExclusion(t *testing.T) {
	f := testutil.NewFixture(t)
	root := filepath.Join(f.Repository, "mod")
	f.WriteFile(t, "repository/mod/file", "payload")
	paths := pathresolve.Resolver{OS: "linux"}
	if _, err := Plan(root, f.Target, config.Settings{}, paths); err == nil || !strings.Contains(err.Error(), "unavailable") {
		t.Fatalf("unavailable module was planned: %v", err)
	}
	if _, err := Plan(root, f.Target, config.Settings{Available: true, Exclude: []string{"../outside"}}, paths); err == nil || !strings.Contains(err.Error(), "cannot escape") {
		t.Fatalf("unsafe exclusion accepted: %v", err)
	}
}
