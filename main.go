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
	cmd.Execute(GitTag, GitCommitHash)
}
