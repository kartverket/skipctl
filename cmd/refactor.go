package cmd

import (
	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/kartverket/skipctl/pkg/refactor"
	"github.com/kartverket/skipctl/pkg/utils"
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
	// Store as a list in case we want to extend the functionality
	var manifestFiles []*manifest.Document
	var err error
	var fileNames []string
	// Extract the given file from the path
	fileNames, err = utils.FindManifestFiles(path)
	if err != nil {
		log.Error("Could not find file(s) to refactor", "error", err.Error())
	}
	// Convert to manifest file format
	manifestFiles, err = manifest.FromFiles(fileNames)
	if err != nil {
		log.Error("Could not convert from files to manifest format", "error", err.Error())
	}
	// Check to see if there are manifests
	if len(manifestFiles) == 0 {
		log.Info("No manifests found")
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
