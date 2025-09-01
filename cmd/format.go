package cmd

import (
	"os"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/spf13/cobra"
)

var formatCmd = &cobra.Command{
	Use:   "format",
	Short: "Formats .jsonnet files",
	Long:  "Recursively formats all .jsonnet files in the specified path",
	Args:  cobra.RangeArgs(0, 1),
	Run:   runFormat,
}

func runFormat(_ *cobra.Command, args []string) {
	var rootDirName = "."

	if len(args) > 0 {
		rootDirName = args[0]
	}

	files, err := findFilesWithSuffixes(rootDirName, constants.ManifestSuffixes)

	if err != nil {
		log.Error("Error collecting files",
			"error", err.Error(),
		)
		os.Exit(1)
	}
	if len(files) == 0 {
		log.Info("No manifests found.")
		return
	}

	processor := manifest.NewManfiestProcessor("format")
	processor.ProcessManifests(files, manifest.FormatManifest)
}
func init() {
	manifestCmd.AddCommand(formatCmd)

}
