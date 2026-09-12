package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOverlayBuildUsesFreshTag(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "source with spaces")
	versionDir := filepath.Join(dir, "internal", "version")
	require.NoError(t, os.MkdirAll(versionDir, 0755))
	source := filepath.Join(versionDir, "stamp.go")
	original := "package version\nvar Version string\n"
	files := map[string]string{
		source:                        original,
		filepath.Join(dir, "go.mod"):  "module example.com/fixture\ngo 1.24.2\n",
		filepath.Join(dir, "main.go"): "package main\nimport (\"fmt\"; \"example.com/fixture/internal/version\")\nfunc main() { fmt.Print(version.Version) }\n",
	}
	for path, contents := range files {
		require.NoError(t, os.WriteFile(path, []byte(contents), 0600))
	}
	for _, tag := range []string{"v1.2.3", "v1.2.4"} {
		t.Run(tag, func(t *testing.T) {
			overlay, err := writeOverlay(t.TempDir(), source, tag)
			require.NoError(t, err)
			binary := filepath.Join(t.TempDir(), "server")
			// Naked supplies its own linker flags. The overlay must still apply.
			cmd := exec.Command("go", "build", "-ldflags", "-s -w", "-o", binary, ".")
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "GOFLAGS=\"-overlay="+overlay+"\"", "GOWORK=off")
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "%s", output)
			output, err = exec.Command(binary).CombinedOutput()
			require.NoError(t, err, "%s", output)
			assert.Equal(t, tag, string(output))
			contents, err := os.ReadFile(source)
			require.NoError(t, err)
			assert.Equal(t, original, string(contents), "stamping must not modify the checkout")
		})
	}
}

func TestRunRejectsMissingCommandAndExistingOverlay(t *testing.T) {
	assert.ErrorContains(t, run(nil), "usage:")
	t.Setenv("GOFLAGS", "-overlay=custom.json")
	assert.ErrorContains(t, run([]string{"go", "build"}), "existing GOFLAGS -overlay")
}
