package cmd

import (
	"fmt"
	"strings"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/kartverket/skipctl/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	ref              string
	verbosityLevel   string
	chunkSize        int
	diffOutputFormat string
	kustomizeEnabled bool
)

var diffCmd = &cobra.Command{
	Use:     "diff",
	Aliases: []string{"d"},
	Short:   "Diff manifest",
	Long: fmt.Sprintf(`Recursively diff manifest files in the specified path.



Verbosity levels:
minimal (output only changed lines)
chunk (output changed lines with 3 lines of context above and below)
full (output the entire file)

Kustomize:
To diff kustomize manifests, you need to set the --kustomize flag and supply a path to a dir where the files to diff against are in --ref.
If the --kustomize flag is not set, kustomize manifests will be skipped!

Supported formats are: %s.

Any errors will be printed to stderr. Returns 0 if all input files are diffed
correctly, otherwise return code 1 is used to indicate failure.`,
		strings.Join(constants.ManifestSuffixes, ", ")),
	RunE:         runDiff,
	SilenceUsage: true,
}

func runDiff(cmd *cobra.Command, _ []string) error {
	// if user did not override verbosity flag, set more sensible defaults based on output format
	if !cmd.Flags().Changed("verbosity") {
		switch diffOutputFormat {
		case constants.DiffOutputPretty:
			verbosityLevel = constants.DiffVerbosityFull
		case constants.DiffOutputPatch:
			verbosityLevel = constants.DiffVerbosityChunk
		case constants.DiffOutputJSON:
			verbosityLevel = constants.DiffVerbosityFull
		}
	}

	filenames, err := utils.FindFilesWithSuffixes(path, constants.ManifestSuffixes)
	if err != nil {
		log.Error("Error collecting files", "error", err.Error())
		return err
	}

	// ignore kustomize-files if the --kustomize flag is not set
	if !kustomizeEnabled {
		filenames = utils.ExcludeSuffixes(filenames, []string{constants.ManifestKustomizeYaml, constants.ManifestKustomizeYml})
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

	// Create appropriate source based on ref type
	var source manifest.Source

	isValidDirectoryRef := utils.IsValidDirectoryRef(ref)
	// If kustomize is enabled, ref must be a valid directory
	if kustomizeEnabled && !isValidDirectoryRef {
		refErr := fmt.Errorf("with --kustomize flag, --ref must be a valid directory path: %s", ref)
		log.Error("invalid ref", "error", refErr)
		return refErr
	}

	// Try ref as directory first
	if isValidDirectoryRef {
		source = manifest.NewDirectorySource(ref, path)
	} else {
		if IsValidCommitRef(ref) {
			source = manifest.NewGitSource(ref)
		} else {
			refErr := fmt.Errorf("invalid commit ref %s", ref)
			log.Error("invalid ref", "error", refErr)
			return refErr
		}
	}

	processor := manifest.NewDocumentProcessor()

	out := logging.RawLogger()
	differ := manifest.NewDiffer(source, out, verbosityLevel, diffOutputFormat, chunkSize)

	err = processor.ProcessDocuments(manifestFiles, differ.Diff)

	return err
}

func init() {
	diffCmd.Flags().StringVar(&ref, "ref", "HEAD", "ref to diff against, use commit hash, branch name or a directory path (default HEAD)")
	diffCmd.Flags().StringVar(&verbosityLevel, "verbosity", constants.DiffVerbosityFull, "Include entire file contents with diffs")
	diffCmd.Flags().IntVar(&chunkSize, "chunk-size", constants.DefaultChunkSize, "Number of lines to include above and below a diff line")
	diffCmd.Flags().StringVar(&diffOutputFormat, "diff-format", constants.DiffOutputPretty, "the output format of the diff (default pretty), allowed (pretty | patch | json)")
	diffCmd.Flags().BoolVar(&kustomizeEnabled, "kustomize", false, "enable diffing of kustomize, if this flag is set, the --ref is expected to be a dir to diff agains, if not set kustomize files will be skipped (default false)")
	manifestCmd.AddCommand(diffCmd)
}
