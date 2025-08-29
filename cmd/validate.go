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

	var files []string

	if len(args) > 0 {
		if strings.HasSuffix(args[0], ".jsonnet") {
			files = append(files, args[0])
		}
	} else {
		err := filepath.Walk(pathname, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && (filepath.Ext(path) == ".jsonnet") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			log.Error("Error walking path",
				"pathname", pathname,
				"error", err.Error(),
			)
			os.Exit(1)
		}
	}

	if len(files) == 0 {
		log.Info("No .jsonnet files found.")
		return
	}

	failed := false
	for _, file := range files {
		err := validator.ValidateManifest(file)
		if err != nil {
			log.Error("validation failed",
				"file", file,
				"error", err.Error(),
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
	validator = manifest.NewJsonnetValidator()
	validateCmd.Flags().StringVar(&pathname, "pathname", ".", "pathname to look for .jsonnet files in")
}
