package cmd

import (
	"fmt"
	"strings"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/kartverket/skipctl/pkg/utils"
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
	files, err := utils.FindManifestFiles(path)
	if err != nil {
		log.Error(err.Error())
		return
	}

	processor := manifest.NewManifestProcessor()
	processor.ProcessManifests(files, manifest.FormatManifest)
}
func init() {
	manifestCmd.AddCommand(formatCmd)
}
