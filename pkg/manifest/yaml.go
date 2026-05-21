package manifest

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/diff"
	"github.com/kartverket/skipctl/pkg/logging"
	"go.yaml.in/yaml/v4"
)

var divider = fmt.Sprintf("%s\n", constants.DocumentSeparator)

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
		r.output.Info(divider)
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
	// Render current file
	d.currentBuffer.Reset()
	err := d.currentRenderer.Render(file)
	if err != nil {
		return nil, false, err
	}
	rendered := d.currentBuffer.String()

	// Get git file
	prevFile, err := d.source.GetPreviousDocument(file)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			diffs, hasChanges := diff.CalculateDiff("", rendered)
			return diffs, hasChanges, nil
		}
		return nil, false, err
	}

	d.gitBuffer.Reset()
	err = d.gitRenderer.Render(prevFile)
	if err != nil {
		return nil, false, err
	}
	prevRendered := d.gitBuffer.String()

	diffs, hasChanges := diff.CalculateDiff(prevRendered, rendered)
	return diffs, hasChanges, nil
}
