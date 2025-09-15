package manifest

import (
	"log/slog"

	"github.com/kartverket/skipctl/pkg/logging"
)

type Differ struct {
	renderer  *Renderer
	rawOutput *slog.Logger
	prevHash  string
}

func NewDiffer(prevHash string) *Differ {
	return &Differ{
		renderer:  NewRenderer(),
		rawOutput: logging.RawLogger(),
		prevHash:  prevHash,
	}
}
func (d *Differ) DiffManifest(file *Document) error {
	prevDoc, err := file.FromPrevHash(d.prevHash)
	if err != nil {
		return err
	}
	d.rawOutput.Info(file.Content)
	d.rawOutput.Info(prevDoc.Content)
	return nil
}
