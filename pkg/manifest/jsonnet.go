package manifest

import (
	"bytes"
	"fmt"
	"log/slog"

	"github.com/google/go-jsonnet"
	"github.com/kartverket/skipctl/pkg/diff"
	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/utils"
)

// JsonnetRenderer renders jsonnet files.
type JsonnetRenderer struct {
	output     *slog.Logger
	vm         *jsonnet.VM
	cache      *ImportCache
	sortOutput bool
}

// NewJsonnetRenderer creates a new jsonnet renderer.
func NewJsonnetRenderer(output *slog.Logger, cache *ImportCache, sortOutput bool, ref ...string) *JsonnetRenderer {
	vm := jsonnet.MakeVM()

	if len(ref) > 0 && !utils.IsValidDirectoryRef(ref[0]) {
		vm.Importer(NewGitFileImporter(ref[0], cache))
	} else {
		vm.Importer(NewFileImporter(cache))
	}

	return &JsonnetRenderer{
		output:     output,
		vm:         vm,
		cache:      cache,
		sortOutput: sortOutput,
	}
}

func (r *JsonnetRenderer) Render(file *Document) error {
	var content string

	if r.sortOutput {
		content = fmt.Sprintf(`
local res = (%s);

if std.type(res) == "array" then
  std.sort(res, function(r)
    std.join("/", [
      r.kind,
      if std.objectHas(r, "metadata") && std.objectHas(r.metadata, "namespace")
        then r.metadata.namespace
        else "",
      r.metadata.name,
    ])
  )
else if std.type(res) == "object" then
  if std.objectHas(res, "kind") && res.kind == "List" then
    res + {items: std.sort(res.items, function(r)
	  std.join("/", [
		r.kind,
        if std.objectHas(r, "metadata") && std.objectHas(r.metadata, "namespace")
          then r.metadata.namespace
          else "",
        r.metadata.name,
	  ])
	)}
else
  res
`, file.Content)
	} else {
		content = file.Content
	}

	node, err := jsonnet.SnippetToAST(file.Name, content)
	if err != nil {
		return fmt.Errorf("parse jsonnet %q: %w", file.Name, err)
	}
	result, err := r.vm.Evaluate(node)
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
func NewJsonnetDiffer(source Source, sortOutput bool) *JsonnetDiffer {
	currentBuf := &bytes.Buffer{}
	gitBuf := &bytes.Buffer{}
	cache := NewImportCache()

	currentLogger := logging.NewRawLoggerTo(currentBuf)
	gitLogger := logging.NewRawLoggerTo(gitBuf)

	return &JsonnetDiffer{
		source:          source,
		currentBuffer:   currentBuf,
		gitBuffer:       gitBuf,
		currentRenderer: NewJsonnetRenderer(currentLogger, cache, sortOutput),
		gitRenderer:     NewJsonnetRenderer(gitLogger, cache, sortOutput, source.Reference()),
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
