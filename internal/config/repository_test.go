package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ADIX7/alma2/internal/testutil"
)

func TestLoadRepositories(t *testing.T) {
	f := testutil.NewFixture(t)
	if got, err := LoadRepositories(filepath.Join(f.Root, "missing-config")); err != nil || len(got) != 0 {
		t.Fatalf("missing config = %v, %v", got, err)
	}
	if got, err := LoadRepositories(""); err != nil || len(got) != 0 {
		t.Fatalf("no config root = %v, %v", got, err)
	}
	f.WriteFile(t, "config/alma/repository.json", `{"repositories":[{"name":"legacy"}]}`)
	f.WriteFile(t, "config/alma/repository.yaml", "repositories:\n  - name: alias\n")
	if got, err := LoadRepositories(filepath.Join(f.ConfigBase, "alma")); err != nil || len(got) != 0 {
		t.Fatalf("JSON/alias-only config was loaded: %v, %v", got, err)
	}
	f.WriteFile(t, "config/alma/repository.yml", "repositories:\n  - name: second\n    repositoryPath: '~/second'\n  - name: first\n    linkPath: '%DOCUMENTS%/links'\n")
	got, err := LoadRepositories(filepath.Join(f.ConfigBase, "alma"))
	if err != nil || len(got) != 2 || got[0].Name != "second" || got[1].Name != "first" || got[0].RepositoryPath == nil || *got[0].RepositoryPath != "~/second" || got[0].LinkPath != nil || got[1].RepositoryPath != nil || got[1].LinkPath == nil || *got[1].LinkPath != "%DOCUMENTS%/links" {
		t.Fatalf("ordered repositories = %+v, %v", got, err)
	}
	f.WriteFile(t, "config/alma/repository.yml", "repositories: []\n")
	if got, err := LoadRepositories(filepath.Join(f.ConfigBase, "alma")); err != nil || len(got) != 0 {
		t.Fatalf("empty repositories = %+v, %v", got, err)
	}
	if _, err := os.Stat(filepath.Join(f.ConfigBase, "alma", "repository.json")); err != nil {
		t.Fatalf("loader changed legacy config: %v", err)
	}
}

func TestInvalidRepositoryYAML(t *testing.T) {
	cases := []struct{ name, yaml, message string }{
		{"empty file", "", "parse"},
		{"missing key", "other: []\n", "expected repositories"},
		{"malformed", "repositories: [\n", "parse"},
		{"root sequence", "- name: invalid\n", "expected repositories"},
		{"null list", "repositories: null\n", "null"},
		{"implicit null", "repositories:\n", "null"},
		{"null item", "repositories: [null]\n", "null"},
		{"null name", "repositories: [{name: null}]\n", "null"},
		{"null optional path", "repositories: [{name: ok, linkPath: null}]\n", "null"},
		{"missing name", "repositories: [{repositoryPath: repo}]\n", "name required"},
		{"blank name", "repositories: [{name: '  '}]\n", "name required"},
		{"wrong list type", "repositories: nope\n", "parse"},
		{"wrong name type", "repositories: [{name: [bad]}]\n", "parse"},
		{"wrong path type", "repositories: [{name: ok, linkPath: 123}]\n", "parse"},
		{"unknown root", "repositories: []\nextra: bad\n", "field extra not found"},
		{"unknown entry", "repositories: [{name: ok, typo: bad}]\n", "field typo not found"},
		{"duplicate field", "repositories: []\nrepositories: []\n", "already defined"},
		{"multiple documents", "repositories: []\n---\nrepositories: []\n", "multiple YAML documents"},
		{"backslash path", "repositories: [{name: ok, repositoryPath: 'foo\\bar'}]\n", "must use /"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			f := testutil.NewFixture(t)
			f.WriteFile(t, "config/alma/repository.yml", tt.yaml)
			_, err := LoadRepositories(filepath.Join(f.ConfigBase, "alma"))
			if err == nil || !strings.Contains(err.Error(), tt.message) {
				t.Fatalf("LoadRepositories error = %v; want %q", err, tt.message)
			}
		})
	}
}
