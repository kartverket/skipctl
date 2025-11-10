package manifest

import (
	"fmt"
	"log/slog"

	"github.com/kartverket/skipctl/pkg/constants"
)

// TypeRenderer handles rendering for a specific manifest type.
type TypeRenderer interface {
	Render(file *Document) error
}

// Renderer provides a unified interface for rendering all manifest types.
type Renderer struct {
	jsonnetRenderer   *JsonnetRenderer
	yamlRenderer      *YamlRenderer
	kustomizeRenderer *KustomizeRenderer
}

// NewRenderer creates a new manifest renderer facade.
func NewRenderer(output *slog.Logger) *Renderer {
	cache := NewImportCache()
	return &Renderer{
		jsonnetRenderer:   NewJsonnetRenderer(output, cache),
		yamlRenderer:      NewYamlRenderer(output),
		kustomizeRenderer: NewKustomizeRenderer(output),
	}
}

func (f *Renderer) Render(file *Document) error {
	switch file.Extension {
	case constants.ManifestSuffixJsonnet:
		return f.jsonnetRenderer.Render(file)
	case constants.ManifestSuffixYaml, constants.ManifestSuffixYml:
		return f.yamlRenderer.Render(file)
	case constants.ManifestKustomizeYaml, constants.ManifestKustomizeYml:
		return f.kustomizeRenderer.Render(file)
	default:
		return fmt.Errorf("unsupported file extension %s", file.Extension)
	}
}
