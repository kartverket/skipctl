package manifest

import (
	"bytes"
	"log/slog"

	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/utils"
)

type Differ struct {
	renderer     *Renderer
	rawOutput    *slog.Logger
	ref          string
	verbose      bool
	renderBuffer *bytes.Buffer
	logger       *slog.Logger
}

func NewDiffer(ref string, verbose bool) *Differ {
	buf := &bytes.Buffer{}
	return &Differ{
		renderer:     NewRenderer(logging.NewRawLoggerTo(buf)),
		rawOutput:    logging.RawLogger(),
		logger:       logging.Logger(),
		ref:          ref,
		verbose:      verbose,
		renderBuffer: buf,
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
	err = d.renderer.RenderManifest(prevFile)
	if err != nil {
		return err
	}
	prevRendered := d.renderBuffer.String()

	diff, hasChanges := utils.Diff(prevRendered, rendered)

	if !hasChanges {
		return nil
	}

	d.logger.Info("diff", "file", file.Name, "ref", d.ref)
	d.rawOutput.Info(utils.DiffsToPrettyPrint(diff, d.verbose))

	return nil
}
