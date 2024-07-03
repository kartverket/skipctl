package main

import (
	"log/slog"
	"os"

	"github.com/kartverket/skipctl/cmd"
)

func main() {
	// TODO: Differentiate between interactive (text) and non-interactive (JSON) use
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cmd.Execute(logger)
}
