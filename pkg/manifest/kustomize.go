package manifest

import (
	"bytes"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/kartverket/skipctl/pkg/diff"
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
	source          Source
	currentBuffer   *bytes.Buffer
	prevBuffer      *bytes.Buffer
	currentRenderer *KustomizeRenderer
	prevRenderer    *KustomizeRenderer
}

// NewKustomizeDiffer creates a new kustomize differ.
func NewKustomizeDiffer(source Source) *KustomizeDiffer {
	currentBuf := &bytes.Buffer{}
	prevBuf := &bytes.Buffer{}

	currentLogger := logging.NewRawLoggerTo(currentBuf)
	prevLogger := logging.NewRawLoggerTo(prevBuf)

	return &KustomizeDiffer{
		source:          source,
		currentBuffer:   currentBuf,
		prevBuffer:      prevBuf,
		currentRenderer: NewKustomizeRenderer(currentLogger, false),
		prevRenderer:    NewKustomizeRenderer(prevLogger, false),
	}
}

func (d *KustomizeDiffer) Diff(file *Document) ([]*diff.ManifestDiff, bool, error) {
	// Render current kustomize
	d.currentBuffer.Reset()
	err := d.currentRenderer.Render(file)
	if err != nil {
		return nil, false, err
	}
	rendered := d.currentBuffer.String()

	prevFile, err := d.source.GetPreviousDocument(file)

	if err != nil {
		return nil, false, err
	}
	// Render previous kustomize
	d.prevBuffer.Reset()
	err = d.prevRenderer.Render(prevFile)
	if err != nil {
		return nil, false, err
	}
	prevRendered := d.prevBuffer.String()

	diffs, hasChanges := diff.CalculateDiff(prevRendered, rendered)

	return diffs, hasChanges, nil
}
