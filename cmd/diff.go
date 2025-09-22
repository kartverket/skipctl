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
	ref              string
	verbosityLevel   string
	chunkSize        int
	diffOutputFormat string
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
	var reRef = regexp.MustCompile(`^(HEAD|[0-9a-fA-F]{7,40}|[A-Za-z0-9._/-]+)$`)

	if !reRef.MatchString(ref) {
		refErr := fmt.Errorf("invalid commit ref %s", ref)
		log.Error(refErr.Error())
		return refErr
	}

	processor := manifest.NewDocumentProcessor()

	differ := manifest.NewDiffer(ref, verbosityLevel, diffOutputFormat, chunkSize)

	err = processor.ProcessDocuments(manifestFiles, differ.DiffManifest)
	if err != nil {
		log.Error("processing error", "error", err)
	}
	return err
}

func init() {
	diffCmd.Flags().StringVar(&ref, "ref", "HEAD", "git ref to diff against, use commit hash or branch name (default HEAD)")
	diffCmd.Flags().StringVar(&verbosityLevel, "verbosity", constants.DiffVerbosityFull, "Include entire file contents with diffs")
	diffCmd.Flags().IntVar(&chunkSize, "chunk-size", constants.DefaultChunkSize, "Number of lines to include above and below a diff line")
	diffCmd.Flags().StringVar(&diffOutputFormat, "diff-format", constants.DiffOutputPretty, "the output format of the diff (default pretty), allowed (pretty | patch | json)")
	manifestCmd.AddCommand(diffCmd)
}
