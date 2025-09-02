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
	path      string
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate manifest files",
	Long:  fmt.Sprintf("Recursively validates %s files in the specified path", strings.Join(constants.ManifestSuffixes, ", ")),
	Run:   runValidate,
}

func runValidate(_ *cobra.Command, _ []string) {

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

	processor := manifest.NewManifestProcessor("validate")

	processor.ProcessManifests(
		files,
		validator.ValidateManifest,
	)
}

func init() {
	manifestCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVar(&path, "path", ".", "pathname to look for"+strings.Join(constants.ManifestSuffixes, ", "))
	validator = manifest.NewValidator()
}
