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
	rawOutput  *slog.Logger
	isFirstDoc bool
	vm         *jsonnet.VM
}

func NewRenderer(loggers ...*slog.Logger) *Renderer {
	var logger *slog.Logger
	if len(loggers) > 0 && loggers[0] != nil {
		logger = loggers[0]
	} else {
		logger = logging.RawLogger()
	}
	return &Renderer{
		isFirstDoc: true,

		rawOutput: logger,
	}
}

func (r *Renderer) SetImporter(importer jsonnet.Importer) {
	r.vm.Importer(importer)
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
	result, err := r.vm.EvaluateFile(file.Name)
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
