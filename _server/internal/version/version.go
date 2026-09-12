// Package version identifies the running server build.
package version

import (
	_ "embed"
	"runtime/debug"
	"strings"
)

// Refresh the tag before release builds. The committed file also supports
// building source archives without Git installed.
//go:generate sh -c "git describe --tags --abbrev=0 > VERSION.tmp && mv VERSION.tmp VERSION"

//go:embed VERSION
var tag string

// Version can be set to a release version with go build -ldflags=-X.
// When unset, String uses the embedded Git tag, then Go's build metadata.
var Version string

// String returns the release override, Git tag, or an unstamped build identifier.
func String() string {
	info, _ := debug.ReadBuildInfo()
	return buildVersion(Version, tag, info)
}

func buildVersion(release, tag string, info *debug.BuildInfo) string {
	if release != "" {
		return release
	}
	if tag = strings.TrimSpace(tag); tag != "" {
		return tag
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
