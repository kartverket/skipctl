package cmd

import (
	"fmt"
	"os"

	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/kartverket/skipctl/pkg/refactor"
	"github.com/spf13/cobra"
)

var refactorCmd = &cobra.Command{
	Use:     "refactor <target-file> [context-files...]",
	Aliases: []string{"re", "r"},
	Short:   "Refactor manifest using AI",
	Long: `Refactor a manifest to ArgoKit v2 format using AI (Vertex AI with Gemini).

The first file is the target to refactor. Additional files provide context (libraries, configs, references).
Output is written to '<first-file>.refactored.jsonnet'.`,
	RunE:         runRefactor,
	Args:         cobra.MinimumNArgs(0),
	SilenceUsage: true,
}

func runRefactor(cmd *cobra.Command, args []string) error {
	var filePaths []string

	// Priority: positional arguments > -p flag
	if len(args) > 0 {
		filePaths = args
	} else if path != "." {
		filePaths = []string{path}
	} else {
		log.Error("No files specified for refactoring")
		return fmt.Errorf("please provide file paths as arguments or use the -p flag")
	}

	// Validate that all paths are files, not directories
	for _, filePath := range filePaths {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			log.Error("Could not access path", "path", filePath, "error", err.Error())
			return err
		}

		if fileInfo.IsDir() {
			log.Error("Refactor only works on files, not directories", "path", filePath)
			return fmt.Errorf("path must be a file, not a directory: %s", filePath)
		}
	}

	// Convert to manifest file format
	manifestFiles, err := manifest.FromFiles(filePaths)
	if err != nil {
		log.Error("Could not convert files to manifest format", "error", err.Error())
		return err
	}

	// Check to see if there are manifests
	if len(manifestFiles) == 0 {
		log.Info("No manifests found")
		return nil
	}

	// Pass all files as context together for refactoring
	err = refactor.RefactorManifest(manifestFiles)
	return err
}
func init() {
	refactorCmd.PersistentFlags().StringVarP(&path, "path", "p", ".", "path to file/directory containing manifest")
	rootCmd.AddCommand(refactorCmd)
}
