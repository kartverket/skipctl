package main

import (
	"github.com/kartverket/skipctl/cmd"
	"github.com/kartverket/skipctl/pkg/telemetry"
)

var (
	// Used for communicating version.
	GitTag        = "0.0.0"
	GitCommitHash string
)

func main() {
	defer telemetry.Close()
	if err := cmd.Execute(GitTag, GitCommitHash); err != nil {
		return
	}
}
