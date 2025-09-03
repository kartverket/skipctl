package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/manifest"
	fileutil "github.com/kartverket/skipctl/pkg/util"
	"github.com/spf13/cobra"
)

var (
	renderer *manifest.Renderer
	path     string
)

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render manifest files to stdout",
	Long: fmt.Sprintf(`Recursively validates manifest files in the specified path.

Supported formats are: %s.

Any valid output will be printed raw to stdout, error messages to stderr. Returns 0 if all input is rendered
correctly, otherwise return code 1 is used to indicate failure.`,
		strings.Join(constants.ManifestSuffixes, ", ")),
	Run: runRender,
}

func runRender(_ *cobra.Command, _ []string) {
	files, err := fileutil.FindFilesWithSuffixes(path, constants.ManifestSuffixes)

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

	processor.ProcessManifests(
		files,
		renderer.RenderManifest,
	)
}

func init() {
	manifestCmd.AddCommand(renderCmd)
	renderCmd.Flags().StringVar(&path, "path", ".", "filesystem path to look for manifests")
	renderer = manifest.NewRenderer()
}
