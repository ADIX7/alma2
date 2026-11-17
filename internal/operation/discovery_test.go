package operation

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ADIX7/alma2/internal/testutil"
)

func TestDiscoverModules(t *testing.T) {
	f := testutil.NewFixture(t)
	f.WriteFile(t, "repository/.alma.yml", "platforms: []\n")
	f.WriteFile(t, "repository/deep/.alma.yml", "platforms: [{platform: windows}]\n")
	f.WriteFile(t, "repository/deep/module/.alma.yml", "platforms: []\n")
	f.WriteFile(t, "repository/.hidden/.alma.yml", "platforms: []\n")
	f.WriteFile(t, "repository/legacy/.alma-config.json", `{}`)
	modules, err := DiscoverModules(f.Repository)
	if err != nil || !reflect.DeepEqual(modules, []string{".hidden", "deep"}) {
		t.Fatalf("discovered: %v, %v", modules, err)
	}
	if err := os.Symlink(filepath.Join(f.Repository, "deep"), filepath.Join(f.Repository, "alias")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	modules, err = DiscoverModules(f.Repository)
	if err != nil || !reflect.DeepEqual(modules, []string{".hidden", "deep"}) {
		t.Fatalf("symlinked directory traversed: %v, %v", modules, err)
	}
	if _, err := DiscoverModules(filepath.Join(f.Repository, "alias")); err == nil || !strings.Contains(err.Error(), "symlinked") {
		t.Fatalf("symlinked root: %v", err)
	}
	f.WriteFile(t, "repository/deep/.alma.yml", "platforms: null\n")
	if _, err := DiscoverModules(f.Repository); err == nil || !strings.Contains(err.Error(), "null") {
		t.Fatalf("invalid direct module config: %v", err)
	}
}
