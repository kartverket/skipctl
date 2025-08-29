package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/spf13/cobra"
)

var (
	pathname  string
	validator *manifest.JsonnetValidator

	green = "\033[32m"
	red   = "\033[31m"
	reset = "\033[0m"
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
			fmt.Printf("Error walking the path %q: %v\n", pathname, err)
			os.Exit(1)
		}
	}

	if len(files) == 0 {
		fmt.Println("No .jsonnet files found.")
		return
	}

	failed := false
	for _, file := range files {
		err := validator.ValidateManifest(file)
		if err != nil {
			fmt.Printf("%sInvalid:%s %s\n  Error: %v\n", red, reset, file, err)
			failed = true
		} else {
			fmt.Printf("%sValid:%s   %s\n", green, reset, file)
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
