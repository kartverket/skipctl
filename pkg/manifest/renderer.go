package manifest

import (
	"log/slog"

	"github.com/google/go-jsonnet"
	"go.yaml.in/yaml/v4"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/utils"
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

func (r *Renderer) RenderManifest(file *utils.ManifestFile) error {
	switch file.Extension {
	case constants.ManifestSuffixJsonnet:
		return r.renderJsonnet(file)
	case constants.ManifestSuffixYaml, constants.ManifestSuffixYml:
		return r.renderYaml(file)
	}
	return nil
}

func (r *Renderer) renderJsonnet(file *utils.ManifestFile) error {
	result, err := r.jsonnet.EvaluateAnonymousSnippet(file.Name, file.Content)

	if err != nil {
		return err
	}
	r.rawOutput.Info(result)

	return nil
}

func (r *Renderer) renderYaml(file *utils.ManifestFile) error {
	var output any

	err := yaml.Unmarshal([]byte(file.Content), &output)
	if err != nil {
		return err
	}

	r.rawOutput.Info("---")
	r.rawOutput.Info(file.Content)

	return nil
}
