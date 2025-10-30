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
	renderer       *Renderer
}

func NewDiffer(ref string, verbosityLevel string, outputFormat string, chunkSize int) *Differ {
	buf := &bytes.Buffer{}
	renderer := NewRenderer(logging.NewRawLoggerTo(buf))
	return &Differ{
		rawOutput:      logging.RawLogger(),
		logger:         logging.Logger(),
		ref:            ref,
		verbosityLevel: verbosityLevel,
		chunkSize:      chunkSize,
		outputFormat:   outputFormat,
		renderBuffer:   buf,
		renderer:       renderer,
	}
}

func (d *Differ) DiffManifest(file *Document) error {
	prevFile, err := file.AtRef(d.ref)
	if err != nil {
		return err
	}

	d.renderBuffer.Reset()
	d.renderer.SetDefaultImporter()
	d.renderer.DisableYamlSeparator()
	err = d.renderer.RenderManifest(file)
	if err != nil {
		return err
	}
	rendered := d.renderBuffer.String()
	prevRendered := ""

	if prevFile.Content != "" {
		d.renderBuffer.Reset()
		d.renderer.SetGitImporter(d.ref)
		d.renderer.DisableYamlSeparator()
		err = d.renderer.RenderManifest(prevFile)
		if err != nil {
			return err
		}
		prevRendered = d.renderBuffer.String()
	}

	diffs, hasChanges := diff.LCS(prevRendered, rendered)
	if !hasChanges {
		return nil
	}

	outputDiffs := diff.FilterDiffs(diffs, d.verbosityLevel, d.chunkSize)

	switch d.outputFormat {
	case constants.DiffOutputPretty:
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
