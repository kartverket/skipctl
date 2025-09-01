package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/spf13/cobra"
)

var (
	validator *manifest.Validator
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate manifest files",
	Long:  fmt.Sprintf("Recursively validates %s files in the specified path", strings.Join(constants.ManifestSuffixes, ", ")),
	Args:  cobra.RangeArgs(0, 1),
	Run:   runValidate,
}

func runValidate(_ *cobra.Command, args []string) {
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

	processor := manifest.NewManfiestProcessor("validate")
	validator := manifest.NewValidator()

	processor.ProcessManifests(
		files,
		validator.ValidateManifest,
	)
}

func init() {
	manifestCmd.AddCommand(validateCmd)
	validator = manifest.NewValidator()
}
