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
	ctx             *Context
	jsonnetDiffer   *JsonnetDiffer
	yamlDiffer      *YamlDiffer
	kustomizeDiffer *KustomizeDiffer
	rawOutput       *slog.Logger
	verbosityLevel  string
	chunkSize       int
	outputFormat    string
}

// NewDiffer creates a new manifest differ facade.
func NewDiffer(ctx *Context, rawOutput *slog.Logger, verbosityLevel string, outputFormat string, chunkSize int) *Differ {
	return &Differ{
		ctx:             ctx,
		jsonnetDiffer:   NewJsonnetDiffer(ctx),
		yamlDiffer:      NewYamlDiffer(ctx),
		kustomizeDiffer: NewKustomizeDiffer(ctx),
		rawOutput:       rawOutput,
		verbosityLevel:  verbosityLevel,
		chunkSize:       chunkSize,
		outputFormat:    outputFormat,
	}
}

func (f *Differ) Diff(file *Document) error {
	var diffs []*diff.ManifestDiff
	var hasDiff bool
	var err error

	switch file.Extension {
	case constants.ManifestSuffixJsonnet:
		diffs, hasDiff, err = f.jsonnetDiffer.Diff(file)
	case constants.ManifestSuffixYaml, constants.ManifestSuffixYml:
		diffs, hasDiff, err = f.yamlDiffer.Diff(file)
	case constants.ManifestKustomizeYaml, constants.ManifestKustomizeYml:
		diffs, hasDiff, err = f.kustomizeDiffer.Diff(file)
	default:
		return fmt.Errorf("unsupported file extension %s", file.Extension)
	}

	if err != nil {
		return err
	}
	if !hasDiff {
		return nil
	}

	outputDiffs := diff.FilterDiffs(diffs, f.verbosityLevel, f.chunkSize)

	switch f.outputFormat {
	case constants.DiffOutputPretty:
		f.rawOutput.Info(diff.DiffsToPrettyPrint(outputDiffs))
		return nil
	case constants.DiffOutputPatch:
		f.rawOutput.Info(diff.DiffsToPatch(outputDiffs, file.Name))
		return nil
	case constants.DiffOutputJSON:
		f.rawOutput.Info(diff.DiffsToJSON(outputDiffs, file.Name, f.ctx.Ref()))
		return nil
	default:
		return fmt.Errorf("invalid format diff output format %s", f.outputFormat)
	}
}
