package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/spf13/cobra"
)

var formatCmd = &cobra.Command{
	Use:   "format",
	Short: "Formats .jsonnet files",
	Long:  "Recursively formats all .jsonnet files in the specified path",
	Args:  cobra.ArbitraryArgs,
	Run:   runFormat,
}

func runFormat(_ *cobra.Command, args []string) {
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
		_, err := manifest.FormatManifest(file)
		if err != nil {
			log.Error("formatting failed",
				"file", file,
				"error", err.Error(),
			)
			failed = true
		} else {
			log.Info("formatting succeeded",
				"file", file,
			)
		}
	}

	if failed {
		os.Exit(1)
	}
}
func init() {
	manifestCmd.AddCommand(formatCmd)
	formatCmd.Flags().StringVar(&pathname, "pathname", ".", "pathname for the desired jsonnet file(s)")
}
