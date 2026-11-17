package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ADIX7/alma2/internal/testutil"
)

func TestLoadModuleAndPresence(t *testing.T) {
	f := testutil.NewFixture(t)
	moduleDir := filepath.Join(f.Repository, "fish")
	f.WriteFile(t, "repository/fish/.alma-config.json", `{"platforms":[]}`)
	f.WriteFile(t, "repository/fish/.alma.yaml", "platforms: []\n")
	if got, err := LoadModule(moduleDir); err != nil || got != nil {
		t.Fatalf("JSON/alias-only module loaded: %+v, %v", got, err)
	}
	f.WriteFile(t, "repository/fish/.alma.yml", "platforms:\n  - platform: default\n    links:\n      - source: config/fish\n        target: '%DOCUMENTS%/fish'\n    install: 'echo foo: bar'\n  - platform: linux\n    target: '~'\n    excludeReadme: false\n    exclude: [scratch]\n    configure: |\n      echo first\n      echo second\n")
	got, err := LoadModule(moduleDir)
	if err != nil || got == nil || len(got.Platforms) != 2 {
		t.Fatalf("module = %+v, %v", got, err)
	}
	first, second := got.Platforms[0], got.Platforms[1]
	if first.Selector != "default" || first.Target != nil || first.ExcludeReadme != nil || len(first.Links) != 1 || first.Links[0].Source != "config/fish" || first.Links[0].Target != "%DOCUMENTS%/fish" || first.Install == nil || *first.Install != "echo foo: bar" {
		t.Fatalf("first section lost presence/values: %+v", first)
	}
	if second.Selector != "linux" || second.Target == nil || *second.Target != "~" || second.ExcludeReadme == nil || *second.ExcludeReadme || second.Configure == nil || !strings.Contains(*second.Configure, "\n") || len(second.Exclude) != 1 {
		t.Fatalf("second section lost values: %+v", second)
	}
	if _, err := os.Stat(filepath.Join(f.Target, "fish")); !os.IsNotExist(err) {
		t.Fatalf("parsing created target: %v", err)
	}
	f.WriteFile(t, "repository/fish/.alma.yml", "platforms: []\n")
	if got, err := LoadModule(moduleDir); err != nil || got == nil || len(got.Platforms) != 0 {
		t.Fatalf("empty platforms = %+v, %v", got, err)
	}
}

func TestInvalidModuleYAML(t *testing.T) {
	cases := []struct{ name, yaml, message string }{
		{"empty", "", "parse"},
		{"missing platforms", "other: []\n", "expected platforms"},
		{"null platforms", "platforms: null\n", "null"},
		{"wrong platforms", "platforms: nope\n", "sequence"},
		{"null section", "platforms: [null]\n", "null"},
		{"missing selector", "platforms: [{}]\n", "selector required"},
		{"numeric selector", "platforms: [{platform: 42}]\n", "must be a string"},
		{"null target", "platforms: [{platform: default, target: ~}]\n", "null"},
		{"quoted home", "platforms: [{platform: default, target: '~'}]\n", ""},
		{"percent path", "platforms: [{platform: default, target: '%APPDATA%/fish'}]\n", ""},
		{"bad boolean", "platforms: [{platform: default, excludeReadme: 'false'}]\n", "boolean"},
		{"null boolean", "platforms: [{platform: default, excludeReadme: null}]\n", "null"},
		{"wrong links", "platforms: [{platform: default, links: nope}]\n", "sequence"},
		{"missing source", "platforms: [{platform: default, links: [{target: fish}]}]\n", "source"},
		{"missing link target", "platforms: [{platform: default, links: [{source: fish}]}]\n", "target required"},
		{"null link target", "platforms: [{platform: default, links: [{source: fish, target: null}]}]\n", "null"},
		{"numeric link target", "platforms: [{platform: default, links: [{source: fish, target: 123}]}]\n", "strings"},
		{"escaping source", "platforms: [{platform: default, links: [{source: '../escape', target: fish}]}]\n", "cannot escape"},
		{"absolute source", "platforms: [{platform: default, links: [{source: /tmp/file, target: fish}]}]\n", "cannot escape"},
		{"Windows absolute source", "platforms: [{platform: default, links: [{source: 'C:/file', target: fish}]}]\n", "cannot escape"},
		{"backslash target", "platforms: [{platform: default, target: 'foo\\bar'}]\n", "must use /"},
		{"backslash source", "platforms: [{platform: default, links: [{source: 'foo\\bar', target: fish}]}]\n", "must use /"},
		{"bad exclusion", "platforms: [{platform: default, exclude: ['../escape']}]\n", "cannot escape"},
		{"numeric exclusion", "platforms: [{platform: default, exclude: [123]}]\n", "strings"},
		{"unknown field", "platforms: [{platform: default, typo: bad}]\n", "field typo not found"},
		{"unknown link field", "platforms: [{platform: default, links: [{source: a, target: b, typo: c}]}]\n", "field typo not found"},
		{"duplicate", "platforms: [{platform: default, platform: linux}]\n", "already defined"},
		{"multiple documents", "platforms: []\n---\nplatforms: []\n", "multiple YAML documents"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			f := testutil.NewFixture(t)
			f.WriteFile(t, "repository/mod/.alma.yml", tt.yaml)
			got, err := LoadModule(filepath.Join(f.Repository, "mod"))
			if tt.message == "" {
				if err != nil || got == nil {
					t.Fatalf("valid config rejected: %+v, %v", got, err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.message) {
				t.Fatalf("error = %v; want %q", err, tt.message)
			}
		})
	}
}
