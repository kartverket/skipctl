package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/kartverket/skipctl/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	commitHash string
	verbose    bool
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Diff manifest",
	Long: fmt.Sprintf(`Recursively diff manifest files in the specified path.

Supported formats are: %s.

Any errors will be printed to stderr. Returns 0 if all input files are diffed
correctly, otherwise return code 1 is used to indicate failure.`,
		strings.Join(constants.ManifestSuffixes, ", ")),
	RunE:          runDiff,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func runDiff(_ *cobra.Command, _ []string) error {
	filenames, err := utils.FindFilesWithSuffixes(path, constants.ManifestSuffixes)
	if err != nil {
		log.Error("Error collecting files", "error", err.Error())
		return err
	}
	manifestFiles, merr := manifest.FromFiles(filenames)
	if merr != nil {
		log.Error("Error collecting manifest files from filenames", "error", merr.Error())
		return merr
	}
	if len(manifestFiles) == 0 {
		log.Info("No manifests found.")
		return nil
	}
	var reRef = regexp.MustCompile(`(?i)^(?:HEAD|[0-9a-f]{7,40})$`)

	if !reRef.MatchString(commitHash) {
		hashErr := fmt.Errorf("invalid commit hash %s", commitHash)
		log.Error(hashErr.Error())
		return hashErr
	}

	processor := manifest.NewDocumentProcessor()

	differ := manifest.NewDiffer(commitHash, verbose)

	err = processor.ProcessDocuments(manifestFiles, differ.DiffManifest)
	if err != nil {
		log.Error("processing error", "error", err)
	}
	return err
}

func init() {
	diffCmd.Flags().StringVar(&commitHash, "hash", "HEAD", "Commit hash to diff against (default HEAD)")
	diffCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Include entire file contents with diffs")
	manifestCmd.AddCommand(diffCmd)
}
