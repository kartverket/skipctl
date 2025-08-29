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
	pathname  string
	validator *manifest.Validator
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate manifest files",
	Long:  fmt.Sprintf("`Recursively validates %v files in the specified path.`", strings.Join(constants.Suffixes, ", ")),
	Args:  cobra.ArbitraryArgs,
	Run:   runValidate,
}

func runValidate(_ *cobra.Command, args []string) {
	var rootDirName = pathname

	if len(args) > 0 {
		rootDirName = args[0]
	}

	files, err := findFilesWithSuffixes(rootDirName, constants.Suffixes)

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
	for _, file := range files {
		validationErr := validator.ValidateManifest(file)
		if validationErr != nil {
			log.Error("validation failed",
				"file", file,
				"error", validationErr.Error(),
			)
			failed = true
		} else {
			log.Info("validation succeeded",
				"file", file,
			)
		}
	}

	if failed {
		os.Exit(1)
	}
}

func init() {
	manifestCmd.AddCommand(validateCmd)
	validator = manifest.NewValidator()
	validateCmd.Flags().StringVar(&pathname, "pathname", ".", fmt.Sprintf("pathname to look for %v files in", strings.Join(constants.Suffixes, ", ")))
}
