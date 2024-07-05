package cmd

import (
	"context"
	"os"

	"github.com/kartverket/skipctl/pkg/test"
	"github.com/spf13/cobra"
)

var (
	hostname string
	count    int32
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Perform a ping from a SKIP cluster",
	Run: func(cmd *cobra.Command, args []string) {
		if len(hostname) == 0 {
			log.Error("no hostname provided")
			os.Exit(1)
		}

		t, err := test.NewTester(hostAddr, tls)
		if err != nil {
			log.Error("could not create client", "error", err)
			os.Exit(1)
		}

		res, err := t.Ping(context.Background(), hostname, count, timeout)
		if err != nil {
			log.Error("could not ping", "error", err)
			os.Exit(1)
		}

		log.Info("ping OK", "result", res)
	},
}

func init() {
	testCmd.AddCommand(pingCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pingCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	pingCmd.Flags().StringVar(&hostname, "hostname", "", "hostname to ping")
	pingCmd.Flags().Int32VarP(&count, "count", "c", 10, "number of pings to send")
}
