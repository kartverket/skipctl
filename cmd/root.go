package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/spf13/cobra"
)

var (
	log          *slog.Logger
	debug        bool
	outputFormat string
)

var rootCmd = &cobra.Command{
	Use:     "skipctl",
	Short:   "A tool for interacting with the SKIP platform",
	Version: "See spf13/cobra#943",
}

func Execute(version, hash string) {
	rootCmd.SetVersionTemplate(fmt.Sprintf("skipctl %s (%s)\n", version, hash))
	// TODO: Why is outputFormat always "text"?
	log = logging.ConfigureLogging(outputFormat, debug)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output", "text", `the output format for logs - must either be "text" or "json"`)
}
