package cmd

import (
	"time"

	"github.com/kartverket/skipctl/pkg/server"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serves API access in order to aid debugging",
	Long: `Intended to be run inside a SKIP cluster in order to help product teams
debug various network issues.`,
	Run: func(_ *cobra.Command, _ []string) {
		if err := server.Serve(addr, globalTimeout, idTokenOrg); err != nil {
			log.Error("could not start server", "error", err)
		}
	},
}

var (
	addr          string
	globalTimeout time.Duration
	idTokenOrg    string
)

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVar(&addr, "addr", "0.0.0.0:3514", "Address to listen on")
	serveCmd.Flags().DurationVar(&globalTimeout, "global-timeout", 1*time.Minute, "Max timeout for all probes")
	serveCmd.Flags().StringVar(&idTokenOrg, "id-token-organization", "kartverket.no", "The organization that is present in valid OIDC ID tokens")
}
