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
	rawOutput     *slog.Logger
	vm            *jsonnet.VM
	sharedCache   *ImportCache
	yamlSeparator bool
}

func NewRenderer(loggers ...*slog.Logger) *Renderer {
	var logger *slog.Logger
	if len(loggers) > 0 && loggers[0] != nil {
		logger = loggers[0]
	} else {
		logger = logging.RawLogger()
	}
	return &Renderer{
		rawOutput:     logger,
		vm:            jsonnet.MakeVM(),
		sharedCache:   NewImportCache(),
		yamlSeparator: false,
	}
}
func (r *Renderer) DisableYamlSeparator() {
	r.yamlSeparator = false
}
func (r *Renderer) SetDefaultImporter() {
	r.vm.Importer(NewFileImporter(r.sharedCache))
}
func (r *Renderer) SetGitImporter(ref string) {
	r.vm.Importer(NewGitFileImporter(ref, r.sharedCache))
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
	result, err := r.vm.Evaluate(node)
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
	if r.yamlSeparator {
		r.rawOutput.Info("---\n")
	}
	r.yamlSeparator = true

	r.rawOutput.Info(file.Content)

	return nil
}
