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
	Short: "Perform a network test",
	Long:  `Perform a network test from the perspective of a SKIP cluster.`,
}

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.PersistentFlags().StringVarP(&hostAddr, "addr", "a", "localhost:3514", "The server to use")
	testCmd.PersistentFlags().DurationVarP(&timeout, "timeout", "t", 30*time.Second, "Timeout for network test")
	testCmd.PersistentFlags().BoolVar(&tls, "tls", true, "Whether to use TLS towards the server")

}
