package cmd

import (
	"context"
	"os"

	"github.com/kartverket/skipctl/pkg/test"
	"github.com/spf13/cobra"
)

var (
	pingHostname string
	pingCount    int32
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Perform a ping from a SKIP cluster",
	Run: func(_ *cobra.Command, _ []string) {
		if len(pingHostname) == 0 {
			log.Error("no hostname provided")
			os.Exit(1)
		}

		t, err := test.NewTester(context.Background(), hostAddr, tls)
		if err != nil {
			log.Error("could not create client", "error", err)
			os.Exit(1)
		}

		res, err := t.Ping(context.Background(), pingHostname, pingCount, timeout)
		if err != nil {
			log.Error("could not ping", "error", err)
			os.Exit(1)
		}

		log.Info("ping OK", "result", res)
	},
}

func init() {
	testCmd.AddCommand(pingCmd)

	pingCmd.Flags().StringVar(&pingHostname, "hostname", "", "hostname to ping")
	pingCmd.Flags().Int32VarP(&pingCount, "count", "c", 10, "number of pings to send")
}
