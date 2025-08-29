package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/spf13/cobra"
)

var (
	pathname  string
	validator *manifest.JsonnetValidator
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate .jsonnet files",
	Long:  `Recursively validates all .jsonnet files in the specified path.`,
	Args:  cobra.ArbitraryArgs,
	Run:   runValidate,
}

func runValidate(_ *cobra.Command, args []string) {
	files, err := collectJsonnetFiles(args)
	if err != nil {
		log.Error("Error collecting files",
			"error", err.Error(),
		)
		os.Exit(1)
	}

	if len(files) == 0 {
		log.Info("No .jsonnet files found.")
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

func collectJsonnetFiles(args []string) ([]string, error) {
	if len(args) > 0 && strings.HasSuffix(args[0], ".jsonnet") {
		return []string{args[0]}, nil
	}

	var files []string
	err := filepath.Walk(pathname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".jsonnet" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func init() {
	manifestCmd.AddCommand(validateCmd)
	validator = manifest.NewJsonnetValidator()
	validateCmd.Flags().StringVar(&pathname, "pathname", ".", "pathname to look for .jsonnet files in")
}
