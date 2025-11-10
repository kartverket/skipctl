package manifest

import (
	"bytes"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/kartverket/skipctl/pkg/diff"
	"github.com/kartverket/skipctl/pkg/git"
	"github.com/kartverket/skipctl/pkg/logging"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// KustomizeRenderer renders kustomize files.
type KustomizeRenderer struct {
	output         *slog.Logger
	kustomizer     *krusty.Kustomizer
	fs             filesys.FileSystem
	printDivider   bool
	dividerEnabled bool
}

// NewKustomizeRenderer creates a new kustomize renderer.
func NewKustomizeRenderer(output *slog.Logger, dividerEnabled ...bool) *KustomizeRenderer {
	opts := krusty.MakeDefaultOptions()

	// SKIP kustomize options
	opts.PluginConfig.HelmConfig.Enabled = true
	opts.PluginConfig.HelmConfig.Command = "helm"
	opts.LoadRestrictions = types.LoadRestrictionsNone

	enableDiv := true
	if len(dividerEnabled) > 0 {
		enableDiv = dividerEnabled[0]
	}

	return &KustomizeRenderer{
		output:         output,
		kustomizer:     krusty.MakeKustomizer(opts),
		fs:             filesys.MakeFsOnDisk(),
		printDivider:   false,
		dividerEnabled: enableDiv,
	}
}
func (r *KustomizeRenderer) addDivider() {
	if r.printDivider && r.dividerEnabled {
		r.output.Info("---\n")
	}
	r.printDivider = true
}

func (r *KustomizeRenderer) Render(file *Document, fs ...filesys.FileSystem) error {
	kustomizeDir := filepath.Dir(file.Name)

	var currentFs = r.fs // use the default filesystem, or git filesystem
	if len(fs) > 0 {
		currentFs = fs[0]
	}

	resMap, err := r.kustomizer.Run(currentFs, kustomizeDir)
	if err != nil {
		return fmt.Errorf("kustomize build %q: %w", file.Name, err)
	}
	yaml, yamlErr := resMap.AsYaml()
	if yamlErr != nil {
		return fmt.Errorf("convert kustomize output to yaml %q: %w", file.Name, yamlErr)
	}

	r.addDivider()
	r.output.Info(string(yaml))
	return nil
}

// KustomizeDiffer diffs kustomize files.
type KustomizeDiffer struct {
	ctx             *Context
	currentBuffer   *bytes.Buffer
	gitBuffer       *bytes.Buffer
	currentRenderer *KustomizeRenderer
	gitRenderer     *KustomizeRenderer
}

// NewKustomizeDiffer creates a new kustomize differ.
func NewKustomizeDiffer(ctx *Context) *KustomizeDiffer {
	currentBuf := &bytes.Buffer{}
	gitBuf := &bytes.Buffer{}

	currentLogger := logging.NewRawLoggerTo(currentBuf)
	gitLogger := logging.NewRawLoggerTo(gitBuf)

	return &KustomizeDiffer{
		ctx:             ctx,
		currentBuffer:   currentBuf,
		gitBuffer:       gitBuf,
		currentRenderer: NewKustomizeRenderer(currentLogger, false),
		gitRenderer:     NewKustomizeRenderer(gitLogger, false),
	}
}

func (d *KustomizeDiffer) Diff(file *Document) ([]*diff.ManifestDiff, bool, error) {
	// Render current file
	d.currentBuffer.Reset()
	err := d.currentRenderer.Render(file)
	if err != nil {
		return nil, false, err
	}
	rendered := d.currentBuffer.String()

	// Get the kustomization file at ref
	prevFile, err := file.AtRef(d.ctx.Ref())
	if err != nil {
		return nil, false, err
	}

	relKustomizePath, err := git.RepoRelativePath(file.Name)
	if err != nil {
		return nil, false, err
	}
	prevFile.Name = "/" + relKustomizePath

	// Get git filesystem
	gitFS, err := d.ctx.GitFS()
	if err != nil {
		return nil, false, fmt.Errorf("load git filesystem: %w", err)
	}

	// Render git file with git filesystem
	d.gitBuffer.Reset()
	err = d.gitRenderer.Render(prevFile, gitFS)
	if err != nil {
		return nil, false, err
	}
	prevRendered := d.gitBuffer.String()

	diffs, hasChanges := diff.LCS(prevRendered, rendered)
	return diffs, hasChanges, nil
}
