// Package version contains version information for the application.
// The information includes the tool version, commit hash, build date, and Go version.
// This package is automatically populated by goreleaser during the release process.
package version

import "runtime"

var (
	Version   = "dev"
	Commit    = "none"
	Date      = "unknown"
	GoVersion = runtime.Version()
)
