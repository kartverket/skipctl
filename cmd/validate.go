package cmd

import (
	"fmt"
	"strings"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/kartverket/skipctl/pkg/utils"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate manifest files against well-known Kubernetes schemas",
	Long:  fmt.Sprintf("Recursively validates %s files in the specified path", strings.Join(constants.ManifestSuffixes, ", ")),
	Run:   runValidate,
}

func runValidate(_ *cobra.Command, _ []string) {
	files, err := utils.FindManifestFiles(path)
	if err != nil {
		log.Error(err.Error())
		return
	}

	processor := manifest.NewManifestProcessor()

	processor.ProcessValidationManifests(files, manifest.NewValidator().ValidateManifest)
}

func init() {
	manifestCmd.AddCommand(validateCmd)
}
