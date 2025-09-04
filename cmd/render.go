package cmd

import (
	"fmt"
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
	Run: runRender,
}

func runRender(_ *cobra.Command, _ []string) {
	files, err := utils.FindManifestFiles(path)
	if err != nil {
		log.Error(err.Error())
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
	renderer = manifest.NewRenderer()
}
