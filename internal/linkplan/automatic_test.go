package linkplan

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ADIX7/alma2/internal/testutil"
)

func TestAutomatic(t *testing.T) {
	f := testutil.NewFixture(t)
	f.WriteFile(t, "repository/mod/.hidden", "hidden")
	f.WriteFile(t, "repository/mod/visible", "payload")
	f.WriteFile(t, "repository/mod/deep/inner", "nested")
	if err := os.Mkdir(filepath.Join(f.Repository, "mod", "empty"), 0700); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(f.Repository, "mod")
	got, err := Automatic(root, f.Target)
	want := []Candidate{
		{Source: filepath.Join(root, ".hidden"), Target: filepath.Join(f.Target, ".hidden"), RelativeSource: ".hidden"},
		{Source: filepath.Join(root, "deep", "inner"), Target: filepath.Join(f.Target, "deep", "inner"), RelativeSource: "deep/inner"},
		{Source: filepath.Join(root, "visible"), Target: filepath.Join(f.Target, "visible"), RelativeSource: "visible"},
	}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("candidates = %+v, %v; want %+v", got, err, want)
	}
	if _, err := os.Stat(filepath.Join(f.Target, "deep")); !os.IsNotExist(err) {
		t.Fatalf("planning created target directories: %v", err)
	}
	if _, err := Automatic(filepath.Join(f.Repository, "missing"), f.Target); err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing root accepted: %v", err)
	}
}

func TestAutomaticSkipsSymlinks(t *testing.T) {
	f := testutil.NewFixture(t)
	root := filepath.Join(f.Repository, "mod")
	f.WriteFile(t, "repository/mod/real/file", "payload")
	f.WriteFile(t, "target/external", "do not link")
	if err := os.Symlink(filepath.Join(f.Target, "external"), filepath.Join(root, "file-alias")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	if err := os.Symlink(filepath.Join(root, "real"), filepath.Join(root, "dir-alias")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	if err := os.Symlink(filepath.Join(f.Target, "missing"), filepath.Join(root, "broken")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	got, err := Automatic(root, f.Target)
	if err != nil || !reflect.DeepEqual(got, []Candidate{{Source: filepath.Join(root, "real", "file"), Target: filepath.Join(f.Target, "real", "file"), RelativeSource: "real/file"}}) {
		t.Fatalf("symlink candidates: %+v, %v", got, err)
	}
	if err := os.Symlink(root, filepath.Join(f.Repository, "root-alias")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	if _, err := Automatic(filepath.Join(f.Repository, "root-alias"), f.Target); err == nil || !strings.Contains(err.Error(), "not a real directory") {
		t.Fatalf("symlink root accepted: %v", err)
	}
}
