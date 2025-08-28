package cmd

import (
	"github.com/spf13/cobra"
)

var manifestCmd = &cobra.Command{
	Use:   "manifest",
	Short: "Work with manifests",
	Long:  `Commands for working with manifests.`,
}

func init() {
	rootCmd.AddCommand(manifestCmd)
}
