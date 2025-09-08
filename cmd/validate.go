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
	tempDir     string
	validateCmd = &cobra.Command{
		Use:   "validate",
		Short: "Validate manifest files against well-known Kubernetes schemas",
		Long:  fmt.Sprintf("Recursively validates %s files in the specified path", strings.Join(constants.ManifestSuffixes, ", ")),
		RunE:  runValidate,
		// SilenceErrors and SilenceUsage are set to true to prevent Cobra from printing errors and usage messages automatically.
		// This allows for custom error handling and logging within the command's execution logic.
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func runValidate(_ *cobra.Command, _ []string) error {
	var err error
	tempDir, err = utils.CreateTempDirectory("schemas")
	if err != nil {
		log.Error("could not create temporary directory for schema files", "error", err)
		return err
	}

	files, err := utils.FindManifestFiles(path)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	processor := manifest.NewManifestDocumentProcessor()
	err = processor.ProcessValidationManifests(files, manifest.NewValidator(tempDir).ValidateManifest)

	// Cobra does not call the PostRun or PersistentPostRun functions if the program exits with an error (os.Exit(>0))
	// Reported in https://github.com/spf13/cobra/issues/1893
	//
	// Similarly, PostRun and PersistentPostRun are not called if RunE returns an error. Therefore we manually intervene
	// to call the cleanUp function using `cobra.OnFinalize`. This ensures that the cleanUp function is always executed
	// when the command  completes, regardless of whether it completes successfully or with an error.
	cobra.OnFinalize(cleanUp)

	return err
}

func init() {
	manifestCmd.AddCommand(validateCmd)
}

func cleanUp() {
	err := os.RemoveAll(tempDir)
	if err != nil {
		log.Error("could not remove temporary directory", "dir", tempDir, "error", err)
	}
}
