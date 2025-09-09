package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/kartverket/skipctl/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	renderer *manifest.Renderer
)

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render manifest files to stdout",
	Long: fmt.Sprintf(`Recursively validates manifest files in the specified path.

Supported formats are: %s.

Any valid output will be printed raw to stdout, error messages to stderr. Returns 0 if all input is rendered
correctly, otherwise return code 1 is used to indicate failure.`,
		strings.Join(constants.ManifestSuffixes, ", ")),
	Run:  runRender,
	Args: cobra.RangeArgs(0, 1),
}

func runRender(_ *cobra.Command, args []string) {
	var manifestFiles []*manifest.Document
	var err error

	if isStdin(args) {
		manifestFiles, err = manifest.FromStdin()
	} else {
		var filenames []string
		filenames, err = utils.FindFilesWithSuffixes(path, constants.ManifestSuffixes)
		if err == nil {
			manifestFiles = manifest.FromFiles(filenames)
		}
	}

	if err != nil {
		log.Error("Error collecting files", "error", err.Error())
		os.Exit(1)
	}
	if len(manifestFiles) == 0 {
		log.Info("No manifests found.")
		return
	}

	processor := manifest.NewDocumentProcessor()

	processor.ProcessDocuments(
		manifestFiles,
		renderer.RenderManifest,
	)
}

func init() {
	manifestCmd.AddCommand(renderCmd)
	renderer = manifest.NewRenderer()
}
