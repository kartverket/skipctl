package cmd

import (
	"context"
	"os"

	"github.com/kartverket/skipctl/pkg/test"
	"github.com/spf13/cobra"
)

var (
	probeHostname string
	probePort     int32
)

var portProbeCmd = &cobra.Command{
	Use:   "probe",
	Short: "Check whether a TCP port is open from a SKIP cluster",
	Run: func(_ *cobra.Command, _ []string) {
		if len(probeHostname) == 0 {
			log.Error("no hostname provided")
			os.Exit(1)
		}
		if probePort == 0 {
			log.Error("no port provided")
			os.Exit(1)
		}

		t, err := test.NewTester(context.Background(), hostAddr, tls)
		if err != nil {
			log.Error("could not create client", "error", err)
			os.Exit(1)
		}

		res, err := t.PortProbe(context.Background(), probeHostname, probePort, timeout)
		if err != nil {
			log.Error("could not probe", "error", err)
			os.Exit(1)
		}

		log.Info("probe finished", "portOpen", res.GetOpen())
	},
}

func init() {
	testCmd.AddCommand(portProbeCmd)

	portProbeCmd.Flags().StringVar(&probeHostname, "hostname", "", "hostname to probe")
	portProbeCmd.Flags().Int32VarP(&probePort, "port", "p", 0, "the port to check")
}
