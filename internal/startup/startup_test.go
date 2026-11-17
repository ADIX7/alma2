package startup

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ADIX7/alma2/internal/testutil"
)

func TestWorkDirectoryPrecedence(t *testing.T) {
	f := testutil.NewFixture(t)
	missing := filepath.Join(f.Root, "missing")
	file := f.WriteFile(t, "not-a-directory", "x")
	for _, tt := range []struct {
		name, primary, secondary, want string
		warnings                       int
	}{
		{"primary", f.Repository, f.Target, f.Repository, 0},
		{"secondary", missing, f.Target, f.Target, 1},
		{"file invalid", file, f.Target, f.Target, 1},
		{"invocation fallback", missing, missing, f.Work, 2},
		{"unset", "", "", f.Work, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			env := map[string]string{"ALMA_WORKDIR": tt.primary, "WORKDIR": tt.secondary}
			got := Resolve(Inputs{Env: func(k string) string { return env[k] }, InvocationDir: f.Work})
			if got.WorkDir != tt.want || len(got.Warnings) != tt.warnings {
				t.Fatalf("work=%q, warnings=%v; want %q, %d warnings", got.WorkDir, got.Warnings, tt.want, tt.warnings)
			}
		})
	}
}

func TestConfigurationProviders(t *testing.T) {
	f := testutil.NewFixture(t)
	userConfig := filepath.Join(f.Home, ".config")
	if err := os.Mkdir(userConfig, 0700); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(f.Root, "missing")
	for _, tt := range []struct {
		name       string
		env        map[string]string
		home, want string
	}{
		{"ALMA_CONFIG even without repository.yml", map[string]string{"ALMA_CONFIG": f.ConfigBase, "XDG_CONFIG_HOME": f.Target}, f.Home, f.ConfigBase},
		{"XDG_CONFIG_HOME", map[string]string{"ALMA_CONFIG": missing, "XDG_CONFIG_HOME": f.Target}, f.Home, f.Target},
		{"user config", map[string]string{"XDG_CONFIG_HOME": missing, "ALMA_CONFIG_FALLBACK": f.Repository}, f.Home, userConfig},
		{"existing fallback", map[string]string{"ALMA_CONFIG_FALLBACK": f.Repository}, "", f.Repository},
		{"nonexistent fallback retained", map[string]string{"ALMA_CONFIG_FALLBACK": missing}, "", missing},
		{"no root", nil, "", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := Resolve(Inputs{Env: func(k string) string { return tt.env[k] }, InvocationDir: f.Work, Home: tt.home})
			want := ""
			if tt.want != "" {
				want = filepath.Join(tt.want, "alma")
			}
			if got.ConfigRoot != want || len(got.ConfigProviders) != 4 {
				t.Fatalf("config=%q, providers=%v; want %q", got.ConfigRoot, got.ConfigProviders, want)
			}
		})
	}
}

func TestApplicationDataProvidersAndEffects(t *testing.T) {
	f := testutil.NewFixture(t)
	missing := filepath.Join(f.Root, "future-data")
	file := f.WriteFile(t, "not-a-directory", "x")
	for _, tt := range []struct {
		name, primary, local, fallback, want string
	}{
		{"ALMA_APP_DATA", f.DataBase, f.Target, f.Repository, f.DataBase},
		{"local app data", file, f.Target, f.Repository, f.Target},
		{"existing fallback", missing, "", f.Repository, f.Repository},
		{"nonexistent fallback retained", missing, "", missing, missing},
		{"workdir fallback", "", "", "", f.Work},
	} {
		t.Run(tt.name, func(t *testing.T) {
			env := map[string]string{"ALMA_APP_DATA": tt.primary, "ALMA_APP_DATA_FALLBACK": tt.fallback}
			got := Resolve(Inputs{Env: func(k string) string { return env[k] }, InvocationDir: f.Work, LocalData: tt.local})
			if got.AppDataRoot != filepath.Join(tt.want, "alma") || len(got.AppDataProviders) != 3 {
				t.Fatalf("app data=%q, providers=%v", got.AppDataRoot, got.AppDataProviders)
			}
		})
	}
	env := map[string]string{"ALMA_APP_DATA_FALLBACK": missing}
	dirs := Resolve(Inputs{Env: func(k string) string { return env[k] }, InvocationDir: f.Work})
	if err := dirs.Prepare(true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Fatalf("dry run created fallback directory: %v", err)
	}
	if err := dirs.Prepare(false); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(dirs.AppDataRoot); err != nil || !info.IsDir() {
		t.Fatalf("real operation did not create app data: %v", err)
	}
	blocked := Directories{AppDataRoot: filepath.Join(file, "alma")}
	if err := blocked.Prepare(false); err == nil {
		t.Fatal("expected creation error beneath a file")
	}
}

func TestProviderValuesAndNoResolveSideEffects(t *testing.T) {
	f := testutil.NewFixture(t)
	missing := filepath.Join(f.Root, "future")
	env := map[string]string{"ALMA_CONFIG": missing, "ALMA_APP_DATA": missing}
	got := Resolve(Inputs{Env: func(k string) string { return env[k] }, InvocationDir: f.Work, Home: f.Home})
	if !reflect.DeepEqual(got.ConfigProviders, []Provider{
		{"ALMA_CONFIG", missing}, {"XDG_CONFIG_HOME", ""},
		{"user config", filepath.Join(f.Home, ".config")}, {"ALMA_CONFIG_FALLBACK", ""},
	}) || got.AppDataProviders[0] != (Provider{"ALMA_APP_DATA", missing}) {
		t.Fatalf("unexpected provider diagnostics: %+v", got)
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Fatalf("resolve created directory: %v", err)
	}
	if strings.HasPrefix(got.AppDataRoot, f.Home) {
		t.Fatalf("app data unexpectedly uses home: %q", got.AppDataRoot)
	}
}
