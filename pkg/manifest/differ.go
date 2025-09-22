package manifest

import (
	"bytes"
	"log/slog"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/utils"
)

type Differ struct {
	renderer       *Renderer
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
		renderer:       NewRenderer(logging.NewRawLoggerTo(buf)),
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
	prevFile, err := file.FromRef(d.ref)
	if err != nil {
		return err
	}

	// render current manifest to buffer
	d.renderBuffer.Reset()
	err = d.renderer.RenderManifest(file)
	if err != nil {
		return err
	}
	rendered := d.renderBuffer.String()

	// render previous manifest to buffer
	d.renderBuffer.Reset()
	d.renderer.SetImporter(NewGitImporter(d.ref))
	err = d.renderer.RenderManifest(prevFile)
	d.renderer.ResetImporter()
	if err != nil {
		return err
	}
	prevRendered := d.renderBuffer.String()

	diffs, hasChanges := utils.DiffLCS(prevRendered, rendered)

	if !hasChanges {
		return nil
	}

	outputDiffs := utils.FilterDiffs(diffs, d.verbosityLevel, d.chunkSize)

	switch d.outputFormat {
	case constants.DiffOutputPretty:
		d.logger.Info("diff", "file", file.Name, "ref", d.ref)
		d.rawOutput.Info(utils.DiffsToPrettyPrint(outputDiffs))
		return nil
	case constants.DiffOutputPatch:
		d.rawOutput.Info(utils.DiffsToPatch(outputDiffs, file.Name))
		return nil
	case constants.DiffOutputJSON:
		d.rawOutput.Info(utils.DiffsToJSON(outputDiffs, file.Name, d.ref))
		return nil
	default:
		return nil
	}
}
