package manifest

import (
	"bytes"
	"fmt"
	"log/slog"

	"github.com/google/go-jsonnet"
	"github.com/kartverket/skipctl/pkg/diff"
	"github.com/kartverket/skipctl/pkg/logging"
)

// JsonnetRenderer renders jsonnet files.
type JsonnetRenderer struct {
	output *slog.Logger
	vm     *jsonnet.VM
	cache  *ImportCache
}

// NewJsonnetRenderer creates a new jsonnet renderer.
func NewJsonnetRenderer(output *slog.Logger, cache *ImportCache, ref ...string) *JsonnetRenderer {
	vm := jsonnet.MakeVM()

	if len(ref) > 0 {
		vm.Importer(NewGitFileImporter(ref[0], cache))
	} else {
		vm.Importer(NewFileImporter(cache))
	}

	return &JsonnetRenderer{
		output: output,
		vm:     vm,
		cache:  cache,
	}
}

func (r *JsonnetRenderer) Render(file *Document) error {
	node, err := jsonnet.SnippetToAST(file.Name, file.Content)
	if err != nil {
		return fmt.Errorf("parse jsonnet %q: %w", file.Name, err)
	}
	result, err := r.vm.Evaluate(node)
	file.Content = result
	file.Rendered = true
	if err != nil {
		return fmt.Errorf("evaluate jsonnet %q: %w", file.Name, err)
	}

	r.output.Info(result)
	return nil
}

// JsonnetDiffer diffs jsonnet files.
type JsonnetDiffer struct {
	source          Source
	currentBuffer   *bytes.Buffer
	gitBuffer       *bytes.Buffer
	currentRenderer *JsonnetRenderer
	gitRenderer     *JsonnetRenderer
}

// NewJsonnetDiffer creates a new jsonnet differ.
func NewJsonnetDiffer(source Source) *JsonnetDiffer {
	currentBuf := &bytes.Buffer{}
	gitBuf := &bytes.Buffer{}
	cache := NewImportCache()

	currentLogger := logging.NewRawLoggerTo(currentBuf)
	gitLogger := logging.NewRawLoggerTo(gitBuf)

	return &JsonnetDiffer{
		source:          source,
		currentBuffer:   currentBuf,
		gitBuffer:       gitBuf,
		currentRenderer: NewJsonnetRenderer(currentLogger, cache),
		gitRenderer:     NewJsonnetRenderer(gitLogger, cache, source.Reference()),
	}
}

func (d *JsonnetDiffer) Diff(file *Document) ([]*diff.ManifestDiff, bool, error) {
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

	diffs, hasChanges := diff.CalculateDiff(prevRendered, rendered)
	return diffs, hasChanges, nil
}
