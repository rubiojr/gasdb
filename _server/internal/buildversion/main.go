// Command buildversion injects the current Git tag into a Go build command.
// It is invoked from the server module by scripts/with-server-version.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: with-server-version COMMAND [ARGS...]")
	}
	flags := os.Getenv("GOFLAGS")
	if strings.Contains(flags, "-overlay") {
		return fmt.Errorf("version stamping cannot be combined with an existing GOFLAGS -overlay")
	}
	source, err := filepath.Abs("internal/version/stamp.go")
	if err != nil {
		return err
	}
	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("run from the _server module: %w", err)
	}
	output, err := exec.Command("git", "describe", "--tags", "--abbrev=0", "--always").Output()
	if err != nil {
		return fmt.Errorf("reading Git version: %w", err)
	}
	tag := strings.TrimSpace(string(output))
	dir, err := os.MkdirTemp("", "gasdb-build-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	overlay, err := writeOverlay(dir, source, tag)
	if err != nil {
		return err
	}
	fmt.Println("GasDB version: " + tag)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = append(os.Environ(), "GOFLAGS="+strings.TrimSpace(flags+" "+strconv.Quote("-overlay="+overlay)))
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

// The overlay changes only the compiler's view of stamp.go. No checkout files
// are modified, and each build has its own stamp, including concurrent builds.
func writeOverlay(dir, source, tag string) (string, error) {
	stamp := filepath.Join(dir, "stamp.go")
	contents := "package version\n\nvar Version = " + strconv.Quote(tag) + "\n"
	if err := os.WriteFile(stamp, []byte(contents), 0600); err != nil {
		return "", err
	}
	data, err := json.Marshal(struct {
		Replace map[string]string
	}{Replace: map[string]string{source: stamp}})
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "overlay.json")
	return path, os.WriteFile(path, data, 0600)
}
