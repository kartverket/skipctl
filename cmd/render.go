package cmd

import (
	"fmt"
	"strings"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/kartverket/skipctl/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	renderer   *manifest.Renderer
	renderPath string
)

var renderCmd = &cobra.Command{
	Use:     "render [path]",
	Aliases: []string{"r"},
	Short:   "Render manifest files to stdout",
	Long: fmt.Sprintf(`Recursively validates manifest files in the specified path.

Supported formats are: %s.

Any valid output will be printed raw to stdout, error messages to stderr. Returns 0 if all input is rendered
correctly, otherwise return code 1 is used to indicate failure.`,
		strings.Join(constants.ManifestSuffixes, ", ")),
	RunE:         runRender,
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
}

func runRender(_ *cobra.Command, args []string) error {
	manifestFiles, err := determineDocuments(renderPath, args, constants.ManifestSuffixes)

	if err != nil {
		log.Error("Error collecting files", "error", err.Error())
		return err
	}
	if len(manifestFiles) == 0 {
		log.Info("No manifests found.")
		return nil
	}

	processor := manifest.NewDocumentProcessor()

	err = processor.ProcessDocuments(
		manifestFiles,
		renderer.Render,
	)

	return err
}

func init() {
	manifestCmd.AddCommand(renderCmd)
	renderer = manifest.NewRenderer(logging.RawLogger())
	renderCmd.Flags().StringVarP(&renderPath, "path", "p", "", "path to render (default: current directory)")
}
