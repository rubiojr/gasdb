package version

// Version is injected by scripts/build-server (also with --deploy) using a
// temporary Go build overlay. It can also be set with go build -ldflags=-X.
// Unstamped builds fall back to Go's embedded build metadata.
var Version string
