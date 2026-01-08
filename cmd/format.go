package cmd

import (
	"fmt"
	"strings"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/kartverket/skipctl/pkg/manifest/format"
	"github.com/kartverket/skipctl/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	formatPath string
)

var formatCmd = &cobra.Command{
	Use:     "format [path]",
	Aliases: []string{"f", "fmt"},
	Short:   "Format manifest in place",
	Long: fmt.Sprintf(`Recursively formats manifest files in the specified path.

Supported formats are: %s.

Any errors will be printed to stderr. Returns 0 if all input files are formatted
correctly, otherwise return code 1 is used to indicate failure.`,
		strings.Join(constants.ManifestSuffixes, ", ")),
	RunE:         runFormat,
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
}

func runFormat(_ *cobra.Command, args []string) error {
	manifestFiles, err := determineDocuments(formatPath, args, constants.FmtManifestSuffixes)

	if err != nil {
		log.Error("Error collecting files", "error", err.Error())
		return err
	}
	if len(manifestFiles) == 0 {
		log.Info("No manifests found.")
		return nil
	}

	processor := manifest.NewDocumentProcessor()
	err = processor.ProcessDocuments(manifestFiles, format.Manifest)

	return err
}
func init() {
	manifestCmd.AddCommand(formatCmd)
	formatCmd.Flags().StringVarP(&formatPath, "path", "p", "", "path to format (default: current directory)")
}
