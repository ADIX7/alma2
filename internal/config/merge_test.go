package config

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ADIX7/alma2/internal/testutil"
)

func TestSelectSpecificityAndOverrides(t *testing.T) {
	f := testutil.NewFixture(t)
	f.WriteFile(t, "repository/mod/.alma.yml", `platforms:
  - platform: linux-fedora-x64
    target: arch
    links: [{source: config, target: arch-config}]
  - platform: linux
    target: linux
    excludeReadme: false
    exclude: [linux-only]
    links: [{source: config, target: linux-config}]
  - platform: default
    target: default
    install: default-install
    exclude: [common]
    links: [{source: config, target: initial}, {source: other, target: other-target}, {source: config, target: final-default}]
  - platform: linux-fedora
    install: fedora-install
    excludeReadme: true
    exclude: [fedora-only]
    links: [{source: other, target: fedora-other}]
  - platform: LINUX-FEDORA
    configure: second-tie
    excludeReadme: false
`)
	module, err := LoadModule(filepath.Join(f.Repository, "mod"))
	if err != nil {
		t.Fatal(err)
	}
	got := Select(module, "linux-fedora-x64")
	if !got.Available || got.Target == nil || *got.Target != "arch" || got.Install == nil || *got.Install != "fedora-install" || got.Configure == nil || *got.Configure != "second-tie" || got.ExcludeReadme {
		t.Fatalf("scalar merge: %+v", got)
	}
	if !reflect.DeepEqual(got.Exclude, []string{"common", "linux-only", "fedora-only"}) {
		t.Fatalf("exclusions: %v", got.Exclude)
	}
	if !reflect.DeepEqual(got.Links, []Link{{"config", "arch-config"}, {"other", "fedora-other"}}) {
		t.Fatalf("links: %v", got.Links)
	}
}

func TestSelectApplicability(t *testing.T) {
	for _, tt := range []struct {
		name, selector, identifier string
		available                  bool
	}{
		{"default", "default", "unknown", true},
		{"text prefix", "linux-fed", "linux-fedora-x64", true},
		{"case insensitive", "LiNuX", "linux-fedora-x64", true},
		{"default is exact", "Default", "linux-x64", false},
		{"unmatched", "windows", "linux-x64", false},
		{"no component boundary", "linux-f", "linux-fedora-x64", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := Select(&Module{Platforms: []Platform{{Selector: tt.selector}}}, tt.identifier)
			if got.Available != tt.available || !got.ExcludeReadme {
				t.Fatalf("select = %+v, want available=%v and README excluded", got, tt.available)
			}
		})
	}
	if got := Select(nil, "linux-x64"); !got.Available || got.ExcludeReadme {
		t.Fatalf("unconfigured module = %+v", got)
	}
	if got := Select(&Module{}, "linux-x64"); got.Available || !got.ExcludeReadme {
		t.Fatalf("configured with no sections = %+v", got)
	}
}

func TestSelectEqualSpecificityAndExplicitFalse(t *testing.T) {
	falseValue := false
	first, second := "first", "second"
	module := &Module{Platforms: []Platform{
		{Selector: "linux", Target: &first, ExcludeReadme: &falseValue},
		{Selector: "LINUX", Target: &second},
	}}
	got := Select(module, "linux-x64")
	if got.Target == nil || *got.Target != "second" || got.ExcludeReadme {
		t.Fatalf("declaration order/explicit false: %+v", got)
	}
}
