package cmd

import (
	"log/slog"
	"os"

	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	log          *slog.Logger
	debug        bool
	outputFormat string
)

var rootCmd = &cobra.Command{
	Use:   "skipctl",
	Short: "A tool for interacting with the SKIP platform",
}

func Execute() {
	log = logging.ConfigureLogging(outputFormat, debug)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output", "text", `the output format for logs - must either be "text" or "json"`)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetEnvPrefix("SKIPCTL")
	viper.AutomaticEnv() // read in environment variables that match
}
