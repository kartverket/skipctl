package cmd

import (
	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/ui"
	"github.com/spf13/cobra"
)

var (
	uiDiscoveryHost string
	uiApiServer     string
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Launch the SKIP dashboard TUI",
	RunE: func(_ *cobra.Command, _ []string) error {
		return ui.Run(uiDiscoveryHost, uiApiServer)
	},
}

func init() {
	rootCmd.AddCommand(uiCmd)
	uiCmd.Flags().StringVar(&uiDiscoveryHost, "discovery-host", constants.DefaultDiscoveryServer, "The DNS name to use for API server discovery")
	uiCmd.Flags().StringVar(&uiApiServer, "api-server", "", "The name of the API server to use")
}
