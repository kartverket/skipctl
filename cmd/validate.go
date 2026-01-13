package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/kartverket/skipctl/pkg/manifest/validate"
	"github.com/kartverket/skipctl/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	tempDir      string
	validatePath string
	validateCmd  = &cobra.Command{
		Use:     "validate [path]",
		Aliases: []string{"v"},
		Short:   "Validate manifest files against well-known Kubernetes schemas",
		Long:    fmt.Sprintf("Recursively validates %s files in the specified path", strings.Join(constants.ManifestSuffixes, ", ")),
		RunE:    runValidate,
		Args:    cobra.MaximumNArgs(1),
		// SilenceErrors and SilenceUsage are set to true to prevent Cobra from printing errors and usage messages automatically.
		// This allows for custom error handling and logging within the command's execution logic.
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func runValidate(_ *cobra.Command, args []string) error {
	manifestFiles, err := determineDocuments(validatePath, args, constants.ManifestSuffixes)

	if err != nil {
		log.Error("Error collecting files", "error", err.Error())
		return err
	}
	if len(manifestFiles) == 0 {
		log.Info("No manifests found.")
		return err
	}

	tempDir, err = utils.CreateTempDirectory("schemas")
	if err != nil {
		log.Error("could not create temporary directory for schema files", "error", err)
		return err
	}

	outputJSON := outputFormat == "json"
	processor := manifest.NewDocumentProcessor()
	validator := validate.NewValidator(tempDir, outputJSON)

	err = processor.ProcessDocuments(manifestFiles, validator.ValidateManifest)

	if err != nil {
		log.Error("processing error", "error", err)
	}

	result := validator.GetResults()
	totalResources := result.GetTotalResources()
	log.InfoContext(logging.DefaultStdoutContext, "validation completed",
		"totalResources", totalResources,
		"valid", result.ValidCount,
		"invalid", result.InvalidCount,
		"errors", result.ErrorCount,
		"skipped", result.SkippedCount,
	)

	// Cobra does not call the PostRun or PersistentPostRun functions if the program exits with an error (os.Exit(>0))
	// Reported in https://github.com/spf13/cobra/issues/1893
	//
	// Similarly, PostRun and PersistentPostRun are not called if RunE returns an error. Therefore we manually intervene
	// to call the cleanUp function using `cobra.OnFinalize`. This ensures that the cleanUp function is always executed
	// when the command  completes, regardless of whether it completes successfully or with an error.
	cobra.OnFinalize(cleanUp)

	if result.HasValidationFailed() {
		return errors.New("validation failed")
	}

	return nil
}

func init() {
	manifestCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVarP(&validatePath, "path", "p", "", "path to validate (default: current directory)")
}

func cleanUp() {
	err := os.RemoveAll(tempDir)
	if err != nil {
		log.Error("could not remove temporary directory", "dir", tempDir, "error", err)
	}
}
