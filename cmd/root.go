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

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initLogging)
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", `the output format for logs - must either be "text" or "json"`)
}

func initLogging() {
	log = logging.ConfigureLogging(outputFormat, debug)
}
