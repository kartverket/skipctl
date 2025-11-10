package manifest

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/diff"
	"github.com/kartverket/skipctl/pkg/git"
	"github.com/kartverket/skipctl/pkg/logging"
)

type Differ struct {
	rawOutput      *slog.Logger
	ref            string
	verbosityLevel string
	chunkSize      int
	outputFormat   string
	renderBuffer   *bytes.Buffer
	logger         *slog.Logger
	renderer       *Renderer
	gitFiles       string
}

func NewDiffer(ref string, verbosityLevel string, outputFormat string, chunkSize int) (*Differ, error) {
	buf := &bytes.Buffer{}
	renderer := NewRenderer(logging.NewRawLoggerTo(buf))

	gitFiles, err := git.GetFilesystemAt(ref)
	if err != nil {
		return nil, err
	}

	return &Differ{
		rawOutput:      logging.RawLogger(),
		logger:         logging.Logger(),
		ref:            ref,
		verbosityLevel: verbosityLevel,
		chunkSize:      chunkSize,
		outputFormat:   outputFormat,
		renderBuffer:   buf,
		renderer:       renderer,
		gitFiles:       gitFiles,
	}, nil
}

func (d *Differ) CleanUpGitFiles() {
	if err := os.RemoveAll(d.gitFiles); err != nil {
		d.logger.Error("failed to clean up git files", "path", d.gitFiles, "error", err)
	}
}

func (d *Differ) DiffManifest(file *Document) error {
	var diffs []*diff.ManifestDiff
	var hasDiff bool
	var err error

	switch file.Extension {
	case constants.ManifestSuffixJsonnet:
		diffs, hasDiff, err = d.diffJsonnet(file)
	case constants.ManifestSuffixYaml, constants.ManifestSuffixYml:
		diffs, hasDiff, err = d.diffYaml(file)
	case constants.ManifestKustomizeYaml, constants.ManifestKustomizeYml:
		diffs, hasDiff, err = d.diffKustomize(file)
	default:
		return fmt.Errorf("unsupported file extension %s", file.Extension)
	}

	if err != nil {
		return err
	}
	if !hasDiff {
		return nil
	}

	outputDiffs := diff.FilterDiffs(diffs, d.verbosityLevel, d.chunkSize)

	switch d.outputFormat {
	case constants.DiffOutputPretty:
		d.rawOutput.Info(diff.DiffsToPrettyPrint(outputDiffs))
		return nil
	case constants.DiffOutputPatch:
		d.rawOutput.Info(diff.DiffsToPatch(outputDiffs, file.Name))
		return nil
	case constants.DiffOutputJSON:
		d.rawOutput.Info(diff.DiffsToJSON(outputDiffs, file.Name, d.ref))
		return nil
	default:
		return fmt.Errorf("invalid format diff output format %s", d.outputFormat)
	}
}

func (d *Differ) diffJsonnet(file *Document) ([]*diff.ManifestDiff, bool, error) {
	prevFile, err := file.AtRef(d.ref)
	if err != nil {
		return nil, false, err
	}

	d.renderBuffer.Reset()
	d.renderer.SetDefaultImporter()
	err = d.renderer.RenderManifest(file)

	if err != nil {
		return nil, false, err
	}
	rendered := d.renderBuffer.String()
	d.renderBuffer.Reset()
	d.renderer.SetGitImporter(d.ref)
	err = d.renderer.RenderManifest(prevFile)

	if err != nil {
		return nil, false, err
	}
	prevRendered := d.renderBuffer.String()

	diffs, hasChanges := diff.LCS(prevRendered, rendered)

	return diffs, hasChanges, nil
}

func (d *Differ) diffYaml(file *Document) ([]*diff.ManifestDiff, bool, error) {
	prevFile, err := file.AtRef(d.ref)
	if err != nil {
		return nil, false, err
	}

	d.renderBuffer.Reset()
	d.renderer.DisableYamlSeparator()
	err = d.renderer.RenderManifest(file)

	if err != nil {
		return nil, false, err
	}
	rendered := d.renderBuffer.String()
	d.renderBuffer.Reset()
	d.renderer.DisableYamlSeparator()
	err = d.renderer.RenderManifest(prevFile)

	if err != nil {
		return nil, false, err
	}
	prevRendered := d.renderBuffer.String()

	diffs, hasChanges := diff.LCS(prevRendered, rendered)

	return diffs, hasChanges, nil
}

func (d *Differ) diffKustomize(file *Document) ([]*diff.ManifestDiff, bool, error) {
	// render the current kustomize file
	d.renderBuffer.Reset()
	renderErr := d.renderer.RenderManifest(file)
	if renderErr != nil {
		return nil, false, renderErr
	}
	rendered := d.renderBuffer.String()
	d.renderBuffer.Reset()

	// get the kustomization file at ref
	prevFile, err := file.AtRef(d.ref)
	if err != nil {
		return nil, false, err
	}

	relKustomizePath, err := git.RepoRelativePath(file.Name)
	if err != nil {
		return nil, false, err
	}
	prevFile.Name = filepath.Join(d.gitFiles, relKustomizePath) // override with filename in git filesystem

	gitRenderErr := d.renderer.RenderManifest(prevFile)
	if gitRenderErr != nil {
		return nil, false, gitRenderErr
	}
	prevRendered := d.renderBuffer.String()

	// diff the outputs
	diffs, hasChanges := diff.LCS(prevRendered, rendered)

	return diffs, hasChanges, nil
}
