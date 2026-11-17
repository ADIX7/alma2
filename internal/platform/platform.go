// Package platform identifies the OS, Linux distribution, and process architecture.
package platform

import (
	"bufio"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// Current returns the legacy-compatible identifier for the running process.
func Current() string {
	return Identifier(runtime.GOOS, runtime.GOARCH, "/etc/os-release")
}

// Identifier accepts an injectable release-file path for fixture tests. An
// absent or unreadable Linux release file simply omits the distribution.
func Identifier(goos, goarch, releaseFile string) string {
	base := ""
	switch goos {
	case "windows":
		base = "windows"
	case "darwin":
		base = "macos"
	case "freebsd":
		base = "freebsd"
	case "linux":
		base = "linux"
		if distro := distribution(releaseFile); distro != "" {
			base += "-" + distro
		}
	default:
		return "unknown"
	}
	return base + "-" + architecture(goarch)
}

func architecture(goarch string) string {
	switch goarch {
	case "386":
		return "x86"
	case "amd64":
		return "x64"
	case "loong64":
		return "loongarch64"
	default:
		return strings.ToLower(goarch)
	}
}

func distribution(releaseFile string) string {
	file, err := os.Open(releaseFile)
	if err != nil {
		return ""
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		key, value, ok := strings.Cut(line, "=")
		if !ok || !strings.EqualFold(key, "ID") {
			continue
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
			value = value[1 : len(value)-1]
		} else if strings.HasPrefix(value, `"`) {
			unquoted, err := strconv.Unquote(value)
			if err != nil {
				return ""
			}
			value = unquoted
		}
		return strings.ToLower(value)
	}
	return ""
}
