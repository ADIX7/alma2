package cli

import "runtime/debug"

func buildVersion() string {
	if build, ok := debug.ReadBuildInfo(); ok && build.Main.Version != "" {
		return build.Main.Version
	}
	return "unknown"
}
