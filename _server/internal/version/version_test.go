package version

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildVersion(t *testing.T) {
	gitInfo := func(revision, modified string) *debug.BuildInfo {
		return &debug.BuildInfo{Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: revision},
			{Key: "vcs.modified", Value: modified},
		}}
	}
	tests := []struct {
		name, release string
		info          *debug.BuildInfo
		want          string
	}{
		{"release override", "v1.2.3", gitInfo("abcdef1234567890", "true"), "v1.2.3"},
		{"release without metadata", "v1.2.3", nil, "v1.2.3"},
		{"clean checkout", "", gitInfo("abcdef1234567890", "false"), "dev-abcdef123456"},
		{"dirty checkout", "", gitInfo("abcdef1234567890", "true"), "dev-abcdef123456-dirty"},
		{"short revision", "", gitInfo("abc", "false"), "dev-abc"},
		{"module release", "", &debug.BuildInfo{Main: debug.Module{Version: "v1.2.3"}}, "v1.2.3"},
		{"development module", "", &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, "dev"},
		{"empty metadata", "", &debug.BuildInfo{}, "dev"},
		{"missing metadata", "", nil, "dev"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, buildVersion(tt.release, tt.info))
		})
	}
}
