package cmd

import (
	"time"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/server"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serves API access in order to aid debugging",
	Long: `Intended to be run inside a SKIP cluster in order to help product teams
debug various connectivity issues.`,
	Run: func(_ *cobra.Command, _ []string) {
		if err := server.Serve(addr, metricsAddr, globalTimeout, idTokenOrg); err != nil {
			log.Error("could not start server", "error", err)
		}
	},
}

var (
	addr          string
	metricsAddr   string
	globalTimeout time.Duration
	idTokenOrg    string
)

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVar(&addr, "addr", "0.0.0.0:3514", "Address to listen on")
	serveCmd.Flags().StringVar(&metricsAddr, "metrics-addr", "0.0.0.0:3515", "Address to listen for metrics on")
	serveCmd.Flags().DurationVar(&globalTimeout, "global-timeout", constants.DefaultServerTestTimeout, "Max timeout for all client probes")
	serveCmd.Flags().StringVar(&idTokenOrg, "id-token-organization", constants.DefaultGoogleOrgID, "The organization that is present in valid OIDC ID tokens")
}
