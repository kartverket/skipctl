package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

var (
	hostAddr string
	timeout  time.Duration
	tls      bool
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Perform a connectivity test",
	Long:  `Perform a connectivity test from the perspective of a SKIP cluster.`,
}

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.PersistentFlags().StringVarP(&hostAddr, "addr", "a", "localhost:3514", "The server to use")
	testCmd.PersistentFlags().DurationVarP(&timeout, "timeout", "t", 10*time.Second, "Timeout for network test") //nolint:lll,mnd // sane default, line length is to be expected
	testCmd.PersistentFlags().BoolVar(&tls, "tls", true, "Whether to use TLS towards the server")
}
