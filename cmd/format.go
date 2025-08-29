package cmd

import (
	"os"

	"github.com/kartverket/skipctl/pkg/constants"
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
