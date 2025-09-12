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
	RunE:          runFormat,
	Args:          cobra.RangeArgs(0, 1),
	SilenceErrors: true,
	SilenceUsage:  true,
}

func runFormat(_ *cobra.Command, args []string) error {
	var manifestFiles []*manifest.Document
	var err error

	if isStdin(args) {
		manifestFiles, err = manifest.FromStdin()
	} else {
		var filenames []string
		filenames, err = utils.FindFilesWithSuffixes(path, constants.ManifestSuffixes)
		if err == nil {
			manifestFiles, err = manifest.FromFiles(filenames)
		}
	}
	if err != nil {
		log.Error("Error collecting files", "error", err.Error())
		return err
	}
	if len(manifestFiles) == 0 {
		log.Info("No manifests found.")
		return nil
	}

	processor := manifest.NewDocumentProcessor()
	err = processor.ProcessDocuments(manifestFiles, manifest.FormatManifest)

	return err
}
func init() {
	manifestCmd.AddCommand(formatCmd)
}
