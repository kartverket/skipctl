package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/spf13/cobra"
)

var formatCmd = &cobra.Command{
	Use:   "format",
	Short: "Format manifest in place",
	Long: fmt.Sprintf(`Recursively formats manifest files in the specified path.

Supported formats are: %s.

Any errors will be printed to stderr. Returns 0 if all input files are formatted
correctly, otherwise return code 1 is used to indicate failure.`,
		strings.Join(constants.ManifestSuffixes, ", ")),
	Run: runFormat,
}

func runFormat(_ *cobra.Command, _ []string) {
	files, err := findFilesWithSuffixes(path, constants.ManifestSuffixes)

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

	processor := manifest.NewManifestProcessor()
	processor.ProcessManifests(files, manifest.FormatManifest)
}
func init() {
	manifestCmd.AddCommand(formatCmd)
	formatCmd.Flags().StringVar(&path, "path", ".", "pathname to look for"+strings.Join(constants.ManifestSuffixes, ", "))
}
