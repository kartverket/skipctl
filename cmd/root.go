package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/kartverket/skipctl/pkg/crd"
	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/telemetry"
	"github.com/spf13/cobra"
)

var (
	log              *slog.Logger
	debug            bool
	outputFormat     string
	disableAnalytics bool
)

var rootCmd = &cobra.Command{
	Use:     "skipctl",
	Short:   "A tool for interacting with the SKIP platform",
	Version: "See spf13/cobra#943",
}

func Execute(version, hash string) error {
	// Wrap commands before execution so wrappers see parsed flag values later.
	instrumentCommands(rootCmd)

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

	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initLogging, initTelemetry)
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", `the output format for logs - must either be "text" or "json"`)
	rootCmd.PersistentFlags().BoolVar(&disableAnalytics, "no-analytics", false, "disable anonymous usage analytics")
}

func initLogging() {
	log = logging.ConfigureLogging(outputFormat, debug)
}

func initTelemetry() {
	telemetry.Init(disableAnalytics)
}

// instrumentCommands wraps each command's Run/RunE to emit telemetry once.
func instrumentCommands(c *cobra.Command) {
	if c.RunE != nil {
		orig := c.RunE
		c.RunE = func(cmd *cobra.Command, args []string) error {
			err := orig(cmd, args)
			cmd.Flags()
			telemetry.CaptureCommand(shortCommandPath(cmd), args, err)
			return err
		}
	}
	for _, child := range c.Commands() {
		instrumentCommands(child)
	}
}
func shortCommandPath(c *cobra.Command) string {
	// Dont capture if only skipctl is invoked
	if c == rootCmd {
		return ""
	}
	parts := strings.Split(c.CommandPath(), " ")
	return strings.Join(parts[1:], " ")
}
