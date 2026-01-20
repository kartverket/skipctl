package render

import (
	"fmt"
	"log/slog"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/manifest"
)

// TypeRenderer handles rendering for a specific manifest type.
type TypeRenderer interface {
	Render(file *manifest.Document) error
}

// Renderer provides a unified interface for rendering all manifest types.
type Renderer struct {
	jsonnetRenderer   *manifest.JsonnetRenderer
	yamlRenderer      *manifest.YamlRenderer
	kustomizeRenderer *manifest.KustomizeRenderer
}

// NewRenderer creates a new manifest renderer facade.
func NewRenderer(output *slog.Logger) *Renderer {
	cache := manifest.NewImportCache()
	return &Renderer{
		jsonnetRenderer:   manifest.NewJsonnetRenderer(output, cache),
		yamlRenderer:      manifest.NewYamlRenderer(output),
		kustomizeRenderer: manifest.NewKustomizeRenderer(output),
	}
}

func (f *Renderer) Render(file *manifest.Document) error {
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
