package manifest

import (
	"fmt"
	"log/slog"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/diff"
)

// TypeDiffer handles diffing for a specific manifest type.
type TypeDiffer interface {
	Diff(file *Document) ([]*diff.ManifestDiff, bool, error)
}

// Differ provides a unified interface for diffing all manifest types.
type Differ struct {
	source          Source
	jsonnetDiffer   *JsonnetDiffer
	yamlDiffer      *YamlDiffer
	kustomizeDiffer *KustomizeDiffer
	rawOutput       *slog.Logger
	verbosityLevel  string
	chunkSize       int
	outputFormat    string
	sortOutput      bool
}

// NewDiffer creates a new manifest differ facade.
func NewDiffer(source Source, rawOutput *slog.Logger, verbosityLevel string, outputFormat string, chunkSize int, sortOutput bool) *Differ {
	return &Differ{
		source:          source,
		jsonnetDiffer:   NewJsonnetDiffer(source, sortOutput),
		yamlDiffer:      NewYamlDiffer(source),
		kustomizeDiffer: NewKustomizeDiffer(source, sortOutput),
		rawOutput:       rawOutput,
		verbosityLevel:  verbosityLevel,
		chunkSize:       chunkSize,
		outputFormat:    outputFormat,
		sortOutput:      sortOutput,
	}
}

func (d *Differ) Diff(file *Document) error {
	var diffs []*diff.ManifestDiff
	var hasDiff bool
	var err error

	switch file.Extension {
	case constants.ManifestSuffixJsonnet:
		diffs, hasDiff, err = d.jsonnetDiffer.Diff(file)
	case constants.ManifestSuffixYaml, constants.ManifestSuffixYml:
		diffs, hasDiff, err = d.yamlDiffer.Diff(file)
	case constants.ManifestKustomizeYaml, constants.ManifestKustomizeYml:
		diffs, hasDiff, err = d.kustomizeDiffer.Diff(file)
	default:
		return fmt.Errorf("unsupported file extension %s", file.Extension)
	}

	if err != nil {
		return err
	}
	if !hasDiff {
		return nil
	}

	outputDiffs := diff.FilterDiffs(diffs, d.verbosityLevel, d.chunkSize)

	switch d.outputFormat {
	case constants.DiffOutputPretty:
		d.rawOutput.Info(diff.DiffsToPrettyPrint(outputDiffs, file.Name))
		return nil
	case constants.DiffOutputPatch:
		d.rawOutput.Info(diff.DiffsToPatch(outputDiffs, file.Name))
		return nil
	case constants.DiffOutputJSON:
		d.rawOutput.Info(diff.DiffsToJSON(outputDiffs, file.Name, d.source.Reference()))
		return nil
	default:
		return fmt.Errorf("invalid format diff output format %s", d.outputFormat)
	}
}

// DiffManifest is an alias for Diff for backward compatibility.
func (d *Differ) DiffManifest(file *Document) error {
	return d.Diff(file)
}
