package cmd

import (
	"github.com/spf13/cobra"
)

var (
	path string
)

var manifestCmd = &cobra.Command{
	Use:   "manifests",
	Short: "Work with manifests",
	Long:  `Commands for working with manifests.`,
}

func init() {
	manifestCmd.PersistentFlags().StringVarP(&path, "path", "p", ".", "path to file/directory containing manifest")
	rootCmd.AddCommand(manifestCmd)
}
