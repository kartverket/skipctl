package cmd

import (
	"time"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/discovery"
	"github.com/spf13/cobra"
)

var (
	apiServer     string
	timeout       time.Duration
	tls           bool
	discoveryHost string
	apiServers    []discovery.APIServer
)

// testCmd represents the test command.
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Perform a connectivity test",
	Long:  `Perform a connectivity test from the perspective of a SKIP cluster.`,
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		servers, err := discovery.DiscoverAPIServers(discoveryHost)
		if err != nil {
			panic(err)
		}
		apiServers = servers
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.PersistentFlags().DurationVarP(&timeout, "timeout", "t", constants.DefaultTestTimeout, "Timeout for network test")
	testCmd.PersistentFlags().BoolVar(&tls, "tls", true, "Whether to use TLS towards the server")
	testCmd.PersistentFlags().StringVar(&discoveryHost, "discovery-host", constants.DefaultDiscoveryServer, "The DNS name to use for API server discovery")
	testCmd.PersistentFlags().StringVar(&apiServer, "api-server", "", "The name of the API server to use")
}
