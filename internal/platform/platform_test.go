package platform

import (
	"path/filepath"
	"testing"

	"github.com/ADIX7/alma2/internal/testutil"
)

func TestIdentifiers(t *testing.T) {
	f := testutil.NewFixture(t)
	release := f.WriteFile(t, "os-release", "NAME=Example\nID=Fedora\nID_LIKE=rhel\n")
	for _, tt := range []struct{ os, arch, want string }{
		{"linux", "amd64", "linux-fedora-x64"},
		{"windows", "386", "windows-x86"},
		{"darwin", "arm64", "macos-arm64"},
		{"freebsd", "arm", "freebsd-arm"},
		{"linux", "s390x", "linux-fedora-s390x"},
		{"linux", "loong64", "linux-fedora-loongarch64"},
		{"linux", "ppc64le", "linux-fedora-ppc64le"},
		{"linux", "wasm", "linux-fedora-wasm"},
		{"linux", "mips64", "linux-fedora-mips64"},
		{"plan9", "amd64", "unknown"},
	} {
		t.Run(tt.os+"-"+tt.arch, func(t *testing.T) {
			if got := Identifier(tt.os, tt.arch, release); got != tt.want {
				t.Fatalf("identifier = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLinuxReleaseFile(t *testing.T) {
	f := testutil.NewFixture(t)
	for _, tt := range []struct{ name, contents, want string }{
		{"quoted", "ID=\"Fedora\"\n", "linux-fedora-x64"},
		{"single quoted", "ID='Alpine'\n", "linux-alpine-x64"},
		{"spaces", " ID = \"Debian\" \n", "linux-x64"},
		{"no ID", "NAME=Example\nID_LIKE=other\n", "linux-x64"},
		{"empty ID", "ID=\n", "linux-x64"},
		{"invalid quote", "ID=\"Fedora\n", "linux-x64"},
		{"case insensitive key", "id=fedora\n", "linux-fedora-x64"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := f.WriteFile(t, "releases/"+tt.name, tt.contents)
			if got := Identifier("linux", "amd64", path); got != tt.want {
				t.Fatalf("identifier = %q, want %q", got, tt.want)
			}
		})
	}
	for _, path := range []string{filepath.Join(f.Root, "missing"), f.Repository} {
		if got := Identifier("linux", "amd64", path); got != "linux-x64" {
			t.Fatalf("unreadable release %q: %q", path, got)
		}
	}
}
