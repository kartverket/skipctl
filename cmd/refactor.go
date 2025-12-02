package cmd

import (
	"fmt"
	"os"

	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/kartverket/skipctl/pkg/refactor"
	"github.com/spf13/cobra"
)

var refactorCmd = &cobra.Command{
	Use:     "refactor",
	Aliases: []string{"re", "r"},
	Short:   "Refactor manifest",
	Long: `One time refactor a manifest in jsonnet to jsonnet ArgoKit v2. 
	Blabla... more info`,
	RunE:         runRefactor,
	Args:         cobra.RangeArgs(0, 1),
	SilenceUsage: true,
}

func runRefactor(cmd *cobra.Command, _ []string) error {
	// Refactor only works on a single file
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Error("Could not access path", "error", err.Error())
		return err
	}

	if fileInfo.IsDir() {
		log.Error("Refactor only works on a single file, not a directory", "path", path)
		return fmt.Errorf("path must be a file, not a directory")
	}

	// Convert to manifest file format
	manifestFiles, err := manifest.FromFiles([]string{path})
	if err != nil {
		log.Error("Could not convert file to manifest format", "error", err.Error())
		return err
	}

	// Check to see if there are manifests
	if len(manifestFiles) == 0 {
		log.Info("No manifest found")
		return nil
	}

	processor := manifest.NewDocumentProcessor()
	err = processor.ProcessDocuments(manifestFiles, refactor.RefactorManifest)
	return err
}
func init() {
	refactorCmd.PersistentFlags().StringVarP(&path, "path", "p", ".", "path to file/directory containing manifest")
	rootCmd.AddCommand(refactorCmd)
}
