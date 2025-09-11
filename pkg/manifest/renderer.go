package manifest

import (
	"fmt"
	"log/slog"

	"github.com/google/go-jsonnet"
	"go.yaml.in/yaml/v4"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/logging"
)

type Renderer struct {
	jsonnet    *jsonnet.VM
	rawOutput  *slog.Logger
	isFirstDoc bool
}

func NewRenderer() *Renderer {
	return &Renderer{
		jsonnet:    jsonnet.MakeVM(),
		rawOutput:  logging.RawLogger(),
		isFirstDoc: true,
	}
}

func (r *Renderer) RenderManifest(file *Document) error {
	switch file.Extension {
	case constants.ManifestSuffixJsonnet:
		return r.renderJsonnet(file)
	case constants.ManifestSuffixYaml, constants.ManifestSuffixYml:
		return r.renderYaml(file)
	}
	return nil
}

func (r *Renderer) renderJsonnet(file *Document) error {
	node, err := jsonnet.SnippetToAST(file.Name, file.Content)
	if err != nil {
		return fmt.Errorf("parse jsonnet %q: %w", file.Name, err)
	}

	result, err := r.jsonnet.Evaluate(node)
	if err != nil {
		return fmt.Errorf("evaluate jsonnet %q: %w", file.Name, err)
	}

	r.rawOutput.Info(result)

	return nil
}

func (r *Renderer) renderYaml(file *Document) error {
	var output any

	err := yaml.Unmarshal([]byte(file.Content), &output)
	if err != nil {
		return err
	}
	if !r.isFirstDoc {
		r.rawOutput.Info("---")
	}
	r.isFirstDoc = false

	r.rawOutput.Info(file.Content)

	return nil
}
