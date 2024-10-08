package cmd

import (
	"context"
	"os"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/test"
	"github.com/spf13/cobra"
)

var (
	pingHostname string
	pingCount    int32
)

var pingCmd = &cobra.Command{
	Use:    "ping",
	Short:  "Perform a ping from a SKIP cluster",
	PreRun: ValidateAPIServerName,
	Run: func(_ *cobra.Command, _ []string) {
		if len(pingHostname) == 0 {
			log.Error("no hostname provided")
			os.Exit(1)
		}

		t, err := test.NewTester(context.Background(), activeAPIServer.Addr, tls)
		if err != nil {
			log.Error("could not create client", "error", err)
			os.Exit(1)
		}

		res, err := t.Ping(context.Background(), pingHostname, pingCount, timeout)
		if err != nil {
			log.Error("could not ping", "error", err)
			os.Exit(1)
		}

		if res.GetPingable() {
			log.Info("successfully pinged", "hostname", pingHostname, "result", res)
		} else {
			log.Info("host not responsive to ping", "hostname", pingHostname, "result", res)
		}
	},
}

func init() {
	testCmd.AddCommand(pingCmd)

	pingCmd.Flags().StringVar(&pingHostname, "hostname", "", "hostname to ping")
	pingCmd.Flags().Int32VarP(&pingCount, "count", "c", constants.DefaultPingCount, "number of pings to send")
}
