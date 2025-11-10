package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/kartverket/skipctl/pkg/crd"
	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/telemetry"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	GitTag        = "0.0.0"
	GitCommitHash string
)

// SetVersionInfo sets version metadata used for CLI version output and telemetry.
func SetVersionInfo(tag, commit string) {
	GitTag = tag
	GitCommitHash = commit
}

var (
	log              *slog.Logger
	collector        telemetry.Collector
	debug            bool
	outputFormat     string
	disableAnalytics bool
)

var rootCmd = &cobra.Command{
	Use:     "skipctl",
	Short:   "A tool for interacting with the SKIP platform",
	Version: "See spf13/cobra#943",
}

func Execute() error {
	defer collector.Close()

	var schemasBuilder strings.Builder
	schemasBuilder.WriteString("Supported schemas:\n")
	schemas, err := crd.ListSchemas()
	if err != nil {
		slog.Error("could not list schemas", "error", err)
		return err
	}
	for _, schema := range schemas {
		schemasBuilder.WriteString(" - ")
		schemasBuilder.WriteString(schema)
		schemasBuilder.WriteString("\n")
	}
	schemasText := schemasBuilder.String()

	rootCmd.SetVersionTemplate(fmt.Sprintf("skipctl %s (%s)\n\n%s", GitTag, GitCommitHash, schemasText))

	executed, err := rootCmd.ExecuteC()
	if executed != nil {
		var flagNames []string
		executed.Flags().Visit(func(f *pflag.Flag) {
			flagNames = append(flagNames, f.Name)
		})

		posArgs := executed.Flags().Args()
		collector.CaptureCommand(shortCommandPath(executed), posArgs, flagNames, err)
	}

	return err
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
	if !disableAnalytics {
		if val, ok := os.LookupEnv("DO_NOT_TRACK"); ok {
			parsed, _ := strconv.ParseBool(val)
			disableAnalytics = parsed
		}
	}

	collector = telemetry.ConfigureCollector(telemetry.Options{
		Debug:            debug,
		DisableAnalytics: disableAnalytics,
		GitVersion:       GitTag,
		GitCommitHash:    GitCommitHash,
	})
}

func shortCommandPath(c *cobra.Command) string {
	if c == rootCmd {
		return rootCmd.DisplayName()
	}
	parts := strings.Split(c.CommandPath(), " ")
	return strings.Join(parts[1:], " ")
}
