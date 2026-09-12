// Package version identifies the running server build.
package version

import "runtime/debug"

// String returns the release override, Git tag, or an unstamped build identifier.
func String() string {
	info, _ := debug.ReadBuildInfo()
	return buildVersion(Version, info)
}

func buildVersion(release string, info *debug.BuildInfo) string {
	if release != "" {
		return release
	}
	return metadataVersion(info)
}

func metadataVersion(info *debug.BuildInfo) string {
	if info == nil {
		return "dev"
	}
	var revision string
	var dirty bool
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			dirty = setting.Value == "true"
		}
	}
	if revision == "" {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
		return "dev"
	}
	result := "dev-" + revision[:min(12, len(revision))]
	if dirty {
		result += "-dirty"
	}
	return result
}
