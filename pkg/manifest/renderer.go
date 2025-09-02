package manifest

import (
	"path/filepath"
	"strings"

	"github.com/google/go-jsonnet"

	"github.com/kartverket/skipctl/pkg/constants"
)

type Renderer struct {
	jsonnet *jsonnet.VM
}

func NewRenderer() *Renderer {
	return &Renderer{
		jsonnet: jsonnet.MakeVM(),
	}
}

func (v *Renderer) RenderManifest(filename string) error {
	extension := strings.ToLower(filepath.Ext(filename))

	switch extension {
	case constants.ManifestSuffixJsonnet:
		return v.renderJsonnet(filename)
	}
	return nil
}

func (v *Renderer) renderJsonnet(filename string) error {
	_, err := v.jsonnet.EvaluateFile(filename)
	return err
}
