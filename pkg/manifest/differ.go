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
	prevHash     string
	verbose      bool
	renderBuffer *bytes.Buffer
	logger       *slog.Logger
}

func NewDiffer(prevHash string, verbose bool) *Differ {
	buf := &bytes.Buffer{}
	return &Differ{
		renderer:     NewRenderer(logging.NewRawLoggerTo(buf)),
		rawOutput:    logging.RawLogger(),
		logger:       logging.Logger(),
		prevHash:     prevHash,
		verbose:      verbose,
		renderBuffer: buf,
	}
}
func (d *Differ) DiffManifest(file *Document) error {
	prevFile, err := file.FromPrevHash(d.prevHash)
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

	diff, hasDiff := utils.Diff(prevRendered, rendered, d.verbose)

	if hasDiff {
		d.logger.Info("diff", "file", file.Name, "commit_hash", d.prevHash)
		d.rawOutput.Info(diff)
	}

	return nil
}
