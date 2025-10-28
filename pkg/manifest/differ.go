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
}

func NewDiffer(ref string, verbosityLevel string, outputFormat string, chunkSize int) *Differ {
	buf := &bytes.Buffer{}
	renderer := NewRenderer(logging.NewRawLoggerTo(buf))
	return &Differ{
		rawOutput:      logging.RawLogger(),
		logger:         logging.Logger(),
		ref:            ref,
		verbosityLevel: verbosityLevel,
		chunkSize:      chunkSize,
		outputFormat:   outputFormat,
		renderBuffer:   buf,
		renderer:       renderer,
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
	// 1. render the current kustomize file
	d.renderBuffer.Reset()
	renderErr := d.renderer.RenderManifest(file)
	if renderErr != nil {
		return nil, false, renderErr
	}
	rendered := d.renderBuffer.String()
	d.renderBuffer.Reset()

	// 2. create a tmp dir and write files form git ref
	kustomizeDir := filepath.Dir(file.Name)
	tmpDir, dirErr := os.MkdirTemp(os.TempDir(), "kustomize-diff-tmp")

	if dirErr != nil {
		return nil, false, dirErr
	}
	defer func() {
		os.RemoveAll(tmpDir)
	}()

	walkErr := filepath.WalkDir(kustomizeDir, func(path string, info os.DirEntry, fileErr error) error {
		if fileErr != nil {
			return fileErr
		}
		if !info.IsDir() {
			content, gitErr := git.GetFileContentAtRef(path, d.ref)
			if gitErr != nil {
				return gitErr
			}

			relPath, _ := filepath.Rel(kustomizeDir, path)
			tmpFilePath := filepath.Join(tmpDir, relPath)
			tmpFileDir := filepath.Dir(tmpFilePath)

			if mkdirErr := os.MkdirAll(tmpFileDir, 0755); mkdirErr != nil {
				return mkdirErr
			}
			if writeErr := os.WriteFile(tmpFilePath, []byte(*content), 0600); writeErr != nil {
				return writeErr
			}
		}
		return nil
	})

	if walkErr != nil {
		return nil, false, walkErr
	}

	// 3. render the git version
	prevFile, err := file.AtRef(d.ref)

	if err != nil {
		return nil, false, err
	}

	prevFile.Name = filepath.Join(tmpDir, "kustomization.yaml")
	gitRenderErr := d.renderer.RenderManifest(prevFile)
	if gitRenderErr != nil {
		return nil, false, gitRenderErr
	}

	prevRendered := d.renderBuffer.String()

	diffs, hasChanges := diff.LCS(prevRendered, rendered)

	return diffs, hasChanges, nil
}
