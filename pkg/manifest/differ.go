package manifest

import (
	"bytes"
	"fmt"
	"log/slog"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/diff"
	"github.com/kartverket/skipctl/pkg/logging"
)

type Differ struct {
	rawOutput      *slog.Logger
	ref            string
	verbosityLevel string
	chunkSize      int
	outputFormat   string
	renderBuffer   *bytes.Buffer
	logger         *slog.Logger
}

func NewDiffer(ref string, verbosityLevel string, outputFormat string, chunkSize int) *Differ {
	buf := &bytes.Buffer{}
	return &Differ{
		rawOutput:      logging.RawLogger(),
		logger:         logging.Logger(),
		ref:            ref,
		verbosityLevel: verbosityLevel,
		chunkSize:      chunkSize,
		outputFormat:   outputFormat,
		renderBuffer:   buf,
	}
}

func (d *Differ) DiffManifest(file *Document) error {
	prevFile, err := file.AtRef(d.ref)
	if err != nil {
		return err
	}

	d.renderBuffer.Reset()
	currentRenderer := NewRenderer(logging.NewRawLoggerTo(d.renderBuffer))
	currentRenderer.SetImporter(NewCachingFileImporter())
	err = currentRenderer.RenderManifest(file)
	if err != nil {
		return err
	}
	rendered := d.renderBuffer.String()

	// Create fresh VM and renderer for previous file
	d.renderBuffer.Reset()
	prevRenderer := NewRenderer(logging.NewRawLoggerTo(d.renderBuffer))
	prevRenderer.SetImporter(NewGitImporter(d.ref))
	err = prevRenderer.RenderManifest(prevFile)
	if err != nil {
		return err
	}
	prevRendered := d.renderBuffer.String()

	diffs, hasChanges := diff.LCS(prevRendered, rendered)

	if !hasChanges {
		return nil
	}

	outputDiffs := diff.FilterDiffs(diffs, d.verbosityLevel, d.chunkSize)

	switch d.outputFormat {
	case constants.DiffOutputPretty:
		d.logger.Info("diff", "file", file.Name, "ref", d.ref)
		d.rawOutput.Info(diff.DiffsToPrettyPrint(outputDiffs))
		return nil
	case constants.DiffOutputPatch:
		d.rawOutput.Info(diff.DiffsToPatch(outputDiffs, file.Name))
		return nil
	case constants.DiffOutputJSON:
		d.rawOutput.Info(diff.DiffsToJSON(outputDiffs, file.Name, d.ref))
		return nil
	default:
		return fmt.Errorf("invalid format diff output format %s", d.outputFormat)
	}
}
