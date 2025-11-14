package manifest

import (
	"bytes"
	"log/slog"

	"github.com/kartverket/skipctl/pkg/diff"
	"github.com/kartverket/skipctl/pkg/logging"
	"go.yaml.in/yaml/v4"
)

// YamlRenderer renders yaml files.
type YamlRenderer struct {
	output         *slog.Logger
	printDivider   bool
	dividerEnabled bool
}

// NewYamlRenderer creates a new yaml renderer.
func NewYamlRenderer(output *slog.Logger, dividerEnabled ...bool) *YamlRenderer {
	enableDiv := true
	if len(dividerEnabled) > 0 {
		enableDiv = dividerEnabled[0]
	}
	return &YamlRenderer{
		output:         output,
		printDivider:   false,
		dividerEnabled: enableDiv,
	}
}

func (r *YamlRenderer) addDivider() {
	if r.printDivider && r.dividerEnabled {
		r.output.Info("---\n")
	}
	r.printDivider = true
}

func (r *YamlRenderer) Render(file *Document) error {
	var output any
	err := yaml.Unmarshal([]byte(file.Content), &output)
	if err != nil {
		return err
	}
	r.addDivider()
	r.output.Info(file.Content)
	return nil
}

// YamlDiffer diffs yaml files.
type YamlDiffer struct {
	source          Source
	currentBuffer   *bytes.Buffer
	gitBuffer       *bytes.Buffer
	currentRenderer *YamlRenderer
	gitRenderer     *YamlRenderer
}

// NewYamlDiffer creates a new yaml differ.
func NewYamlDiffer(source Source) *YamlDiffer {
	currentBuf := &bytes.Buffer{}
	gitBuf := &bytes.Buffer{}

	currentLogger := logging.NewRawLoggerTo(currentBuf)
	gitLogger := logging.NewRawLoggerTo(gitBuf)

	return &YamlDiffer{
		source:          source,
		currentBuffer:   currentBuf,
		gitBuffer:       gitBuf,
		currentRenderer: NewYamlRenderer(currentLogger, false),
		gitRenderer:     NewYamlRenderer(gitLogger, false),
	}
}

func (d *YamlDiffer) Diff(file *Document) ([]*diff.ManifestDiff, bool, error) {
	prevFile, err := d.source.GetPreviousDocument(file)
	if err != nil {
		return nil, false, err
	}

	// Render current file
	d.currentBuffer.Reset()
	err = d.currentRenderer.Render(file)
	if err != nil {
		return nil, false, err
	}
	rendered := d.currentBuffer.String()

	// Render git file
	d.gitBuffer.Reset()
	err = d.gitRenderer.Render(prevFile)
	if err != nil {
		return nil, false, err
	}
	prevRendered := d.gitBuffer.String()

	diffs, hasChanges := diff.LCS(prevRendered, rendered)
	return diffs, hasChanges, nil
}
