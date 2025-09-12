package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/kartverket/skipctl/pkg/crd"
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
	schemasText := "Supported schemas:\n"
	schemas, err := crd.ListSchemas()
	if err != nil {
		slog.Error("could not list schemas", "error", err)
		os.Exit(1)
	}
	for _, schema := range schemas {
		schemasText += fmt.Sprintf(" - %s\n", schema)
	}

	rootCmd.SetVersionTemplate(fmt.Sprintf("skipctl %s (%s)\n\n%s", version, hash, schemasText))

	execErr := rootCmd.Execute()
	if execErr != nil {
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
