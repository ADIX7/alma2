package history

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ADIX7/alma2/internal/linkplan"
	"github.com/ADIX7/alma2/internal/testutil"
	"gopkg.in/yaml.v3"
)

func readRecord(t *testing.T, file string) Record {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var record Record
	if err := yaml.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	return record
}

func TestSavePairsDedupAndOldFilesUntouched(t *testing.T) {
	f := testutil.NewFixture(t)
	dataRoot := filepath.Join(f.DataBase, "alma")
	old := f.WriteFile(t, "data/alma/moduleHashes.txt", "legacy index")
	legacy := f.WriteFile(t, "data/alma/0123456789abcdef.txt", "legacy record")
	source := filepath.Join(f.Repository, "mod")
	target := f.Target
	a, b := filepath.Join(target, "a"), filepath.Join(target, "b")
	path, err := File(dataRoot, source, target)
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(dataRoot, source, target, nil); err != nil {
		t.Fatal(err)
	}
	if got := readRecord(t, path); got.Version != 1 || got.Source != source || got.Target != target || len(got.Destinations) != 0 {
		t.Fatalf("empty record: %+v", got)
	}
	if err := Save(dataRoot, source, target, []linkplan.Candidate{{Target: a}, {Target: b}, {Target: a}}); err != nil {
		t.Fatal(err)
	}
	if err := Save(dataRoot, source, target, []linkplan.Candidate{{Target: b}}); err != nil {
		t.Fatal(err)
	}
	if got := readRecord(t, path); !reflect.DeepEqual(got.Destinations, []string{a, b}) {
		t.Fatalf("dedup: %+v", got)
	}
	another, err := File(dataRoot, source, filepath.Join(target, "other"))
	if err != nil {
		t.Fatal(err)
	}
	if another == path {
		t.Fatal("different target roots share history")
	}
	if err := Save(dataRoot, source, filepath.Join(target, "other"), nil); err != nil {
		t.Fatal(err)
	}
	if got := readRecord(t, another); len(got.Destinations) != 0 {
		t.Fatalf("other pair inherited links: %+v", got)
	}
	for _, file := range []string{old, legacy} {
		contents, err := os.ReadFile(file)
		if err != nil || string(contents) != map[string]string{old: "legacy index", legacy: "legacy record"}[file] {
			t.Fatalf("old metadata changed: %q, %v", contents, err)
		}
	}
}

func TestSaveWithoutDefaultTarget(t *testing.T) {
	f := testutil.NewFixture(t)
	root := filepath.Join(f.DataBase, "alma")
	file, err := File(root, f.Repository, "")
	if err != nil {
		t.Fatal(err)
	}
	cwdFile, err := File(root, f.Repository, ".")
	if err != nil || cwdFile == file {
		t.Fatalf("unset target shares history with cwd: %q, %v", cwdFile, err)
	}
	dest := filepath.Join(f.Target, "mapped")
	if err := Save(root, f.Repository, "", []linkplan.Candidate{{Target: dest}}); err != nil {
		t.Fatal(err)
	}
	if err := Save(root, f.Repository, "", nil); err != nil {
		t.Fatal(err)
	}
	if got := readRecord(t, file); got.Target != "" || !reflect.DeepEqual(got.Destinations, []string{dest}) {
		t.Fatalf("no-default-target history: %+v", got)
	}
}

func TestSaveRejectsCorruptAndUnwritableHistory(t *testing.T) {
	f := testutil.NewFixture(t)
	root := filepath.Join(f.DataBase, "alma")
	source, target := f.Repository, f.Target
	path, err := File(root, source, target)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("version: 2\nsource: mismatch\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Save(root, source, target, nil); err == nil {
		t.Fatal("wrong history identity accepted")
	}
	contents, err := os.ReadFile(path)
	if err != nil || string(contents) != "version: 2\nsource: mismatch\n" {
		t.Fatalf("corrupt history overwritten: %q, %v", contents, err)
	}
	other := filepath.Join(f.Root, "blocked")
	f.WriteFile(t, "blocked/history-v1", "parent is file")
	if err := Save(other, source, target, nil); err == nil {
		t.Fatal("unwritable history accepted")
	}
}
