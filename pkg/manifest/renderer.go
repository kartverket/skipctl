package manifest

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-jsonnet"
	"go.yaml.in/yaml/v4"

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

func (r *Renderer) RenderManifest(filename string) error {
	extension := strings.ToLower(filepath.Ext(filename))

	switch extension {
	case constants.ManifestSuffixJsonnet:
		return r.renderJsonnet(filename)
	case constants.ManifestSuffixYaml:
		return r.renderYaml(filename)
	}

	return nil
}

func (r *Renderer) renderJsonnet(filename string) error {
	result, err := r.jsonnet.EvaluateFile(filename)

	if err != nil {
		return err
	}
	r.rawOutput.Info(result)

	return nil
}

func (r *Renderer) renderYaml(filename string) error {
	fileContents, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var output any

	return yaml.Unmarshal(fileContents, &output)
}
