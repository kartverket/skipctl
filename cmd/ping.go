package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/test"
	"github.com/spf13/cobra"
)

var (
	pingHostname string
	pingCount    int32
)

var pingCmd = &cobra.Command{
	Use:     "ping",
	Short:   "Perform a ping from a SKIP cluster",
	PreRunE: ValidateAPIServerName,
	RunE: func(_ *cobra.Command, _ []string) error {
		if len(pingHostname) == 0 {
			return errors.New("no hostname provided")
		}

		t, err := test.NewTester(context.Background(), activeAPIServer.Addr, tls)
		if err != nil {
			return fmt.Errorf("could not create client: %w", err)
		}

		res, err := t.Ping(context.Background(), pingHostname, pingCount, timeout)
		if err != nil {
			return fmt.Errorf("could not ping: %w", err)
		}

		if res.GetPingable() {
			log.Info("successfully pinged", constants.HostnameKey, pingHostname, "result", res)
		} else {
			log.Info("host not responsive to ping", constants.HostnameKey, pingHostname, "result", res)
		}

		return nil
	},
	SilenceUsage: true,
}

func init() {
	testCmd.AddCommand(pingCmd)

	pingCmd.Flags().StringVar(&pingHostname, "hostname", "", "hostname to ping")
	pingCmd.Flags().Int32VarP(&pingCount, "count", "c", constants.DefaultPingCount, "number of pings to send")
}
