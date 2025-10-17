package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/kartverket/skipctl/pkg/test"
	"github.com/spf13/cobra"
)

var (
	probeHostname string
	probePort     int32
)

var portProbeCmd = &cobra.Command{
	Use:     "probe",
	Short:   "Check whether a TCP port is open from a SKIP cluster",
	PreRunE: ValidateAPIServerName,
	RunE: func(_ *cobra.Command, _ []string) error {
		if len(probeHostname) == 0 {
			return errors.New("no hostname provided")
		}
		if probePort == 0 {
			return errors.New("no port provided")
		}

		t, err := test.NewTester(context.Background(), activeAPIServer.Addr, tls)
		if err != nil {
			return fmt.Errorf("could not create client: %v", err)
		}

		res, err := t.PortProbe(context.Background(), probeHostname, probePort, timeout)
		if err != nil {
			return fmt.Errorf("could not probe: %v", err)
		}

		log.Info("probe finished", "portOpen", res.GetOpen())
		return nil
	},
}

func init() {
	testCmd.AddCommand(portProbeCmd)

	portProbeCmd.Flags().StringVar(&probeHostname, "hostname", "", "hostname to probe")
	portProbeCmd.Flags().Int32VarP(&probePort, "port", "p", 0, "the port to check")
}
