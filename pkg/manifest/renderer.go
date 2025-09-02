package manifest

import (
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/google/go-jsonnet"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/logging"
)

type Renderer struct {
	jsonnet   *jsonnet.VM
	rawOutput *slog.Logger
}

func NewRenderer() *Renderer {
	return &Renderer{
		jsonnet:   jsonnet.MakeVM(),
		rawOutput: logging.RawLogger(),
	}
}

func (v *Renderer) RenderManifest(filename string) error {
	extension := strings.ToLower(filepath.Ext(filename))

	switch extension { //nolint:gocritic // singleCaseSwitch: this is intentional
	case constants.ManifestSuffixJsonnet:
		return v.renderJsonnet(filename)
	}
	return nil
}

func (v *Renderer) renderJsonnet(filename string) error {
	result, err := v.jsonnet.EvaluateFile(filename)

	if err != nil {
		return err
	}
	v.rawOutput.Info(result)

	return nil
}
