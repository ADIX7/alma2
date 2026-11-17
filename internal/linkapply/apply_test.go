package linkapply

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ADIX7/alma2/internal/linkplan"
	"github.com/ADIX7/alma2/internal/testutil"
)

func TestApplyCreatesLinksAndPreservesConflicts(t *testing.T) {
	f := testutil.NewFixture(t)
	source := f.WriteFile(t, "repository/mod/file", "payload")
	directory := filepath.Join(f.Repository, "mod")
	existing := f.WriteFile(t, "target/existing", "keep")
	candidates := []linkplan.Candidate{
		{Source: source, Target: filepath.Join(f.Target, "nested", "file")},
		{Source: directory, Target: filepath.Join(f.Target, "folder"), Directory: true},
		{Source: source, Target: existing},
	}
	var output bytes.Buffer
	created, err := Apply(candidates, &output)
	if err != nil || len(created) != 2 || created[0].Target != candidates[0].Target || created[1].Target != candidates[1].Target || !strings.Contains(output.String(), "Item already exists") {
		t.Fatalf("result: %+v, %v, %q", created, err, output.String())
	}
	for _, candidate := range candidates[:2] {
		info, err := os.Lstat(candidate.Target)
		if err != nil || info.Mode()&os.ModeSymlink == 0 {
			t.Fatalf("no symlink at %q: %v", candidate.Target, err)
		}
		linked, err := os.Readlink(candidate.Target)
		if err != nil || linked != candidate.Source {
			t.Fatalf("link target: %q, %v", linked, err)
		}
	}
	content, err := os.ReadFile(existing)
	if err != nil || string(content) != "keep" {
		t.Fatalf("conflict overwritten: %q, %v", content, err)
	}
	output.Reset()
	again, err := Apply(candidates, &output)
	if err != nil || len(again) != 0 {
		t.Fatalf("repeat run: %+v, %v", again, err)
	}
}

func TestApplyPermissionFailureWhenEnforced(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("permission test needs non-root POSIX filesystem enforcement")
	}
	f := testutil.NewFixture(t)
	source := f.WriteFile(t, "repository/mod/file", "payload")
	locked := filepath.Join(f.Target, "locked")
	if err := os.Mkdir(locked, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0700) })
	if err := os.Chmod(locked, 0500); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(locked, "file")
	var output bytes.Buffer
	created, err := Apply([]linkplan.Candidate{{Source: source, Target: target}}, &output)
	if err == nil {
		_ = os.Remove(target)
		t.Skip("filesystem did not enforce directory write restriction")
	}
	if len(created) != 0 {
		t.Fatalf("permission-denied item recorded: %+v", created)
	}
}

func TestApplyContinuesOnFailureAndRejectsChangedSource(t *testing.T) {
	f := testutil.NewFixture(t)
	a := f.WriteFile(t, "repository/mod/a", "a")
	b := f.WriteFile(t, "repository/mod/b", "b")
	f.WriteFile(t, "target/blocker", "not a directory")
	candidates := []linkplan.Candidate{
		{Source: a, Target: filepath.Join(f.Target, "blocker", "failed")},
		{Source: filepath.Join(f.Repository, "mod", "missing"), Target: filepath.Join(f.Target, "missing-source")},
		{Source: b, Target: filepath.Join(f.Target, "good")},
	}
	var output bytes.Buffer
	created, err := Apply(candidates, &output)
	if err == nil || len(created) != 1 || created[0].Target != candidates[2].Target || !strings.Contains(output.String(), "create link") {
		t.Fatalf("partial run: %+v, %v, %q", created, err, output.String())
	}
	info, err := os.Lstat(candidates[2].Target)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("success not retained: %v", err)
	}
	if _, err := os.Lstat(candidates[1].Target); !os.IsNotExist(err) {
		t.Fatalf("missing source created link: %v", err)
	}
	if err := os.Symlink(a, filepath.Join(f.Repository, "mod", "alias")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	_, err = Apply([]linkplan.Candidate{{Source: filepath.Join(f.Repository, "mod", "alias"), Target: filepath.Join(f.Target, "alias")}}, &output)
	if err == nil || !strings.Contains(err.Error(), "source type changed") {
		t.Fatalf("symlinked source accepted: %v", err)
	}
}
