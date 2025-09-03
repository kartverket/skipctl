package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/spf13/cobra"
)

var (
	validator *manifest.Validator
	logger    = logging.RawLogger()
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

	failed := false

	var validCount, invalidCount, errorCount, skippedCount int

	for _, file := range files {
		summary, validationErr := validator.ValidateManifest(file)
		if validationErr != nil {
			logger.Error(fmt.Sprintf("✖ Validation failed for %s:\n %s \n", file, validationErr.Error()))
			failed = true
		}

		validCount += summary.ValidCount
		invalidCount += summary.InvalidCount
		errorCount += summary.ErrorCount
		skippedCount += summary.SkippedCount
	}

	totalResources := validCount + invalidCount + errorCount + skippedCount

	logger.Info(fmt.Sprintf("Summary: %d resources found - Valid: %d, Invalid: %d, Errors: %d, Skipped: %d \n", totalResources, validCount, invalidCount, errorCount, skippedCount))

	if failed {
		os.Exit(1)
	}
}

func init() {
	manifestCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVarP(&path, "path", "p", ".", "filesystem path to look for manifests")
	validator = manifest.NewValidator()
}
