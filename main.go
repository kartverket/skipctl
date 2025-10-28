package main

import (
	"github.com/kartverket/skipctl/cmd"
)

var (
	// Used for communicating version.
	GitTag        = "0.0.0"
	GitCommitHash string
)

func main() {
	cmd.SetVersionInfo(GitTag, GitCommitHash)

	if err := cmd.Execute(); err != nil {
		return
	}
}
