package manifest

import (
	"bytes"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/diff"
	"github.com/kartverket/skipctl/pkg/logging"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	k8syaml "sigs.k8s.io/yaml"
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
	annotateKustomizeDiffsWithResourceMeta(diffs, prevRendered, rendered)

	return diffs, hasChanges, nil
}

type kustomizeResourceMeta struct {
	Kind       string
	APIVersion string
	Name       string
	Namespace  string
}

func annotateKustomizeDiffsWithResourceMeta(diffs []*diff.ManifestDiff, prevRendered, rendered string) {
	oldMap := buildLineToResourceMetaMap(prevRendered)
	newMap := buildLineToResourceMetaMap(rendered)

	for _, d := range diffs {
		var meta kustomizeResourceMeta
		switch d.Type {
		case constants.Insertion:
			meta = newMap[d.NewLine]
		case constants.Deletion:
			meta = oldMap[d.OldLine]
		default:
			if m, ok := newMap[d.NewLine]; ok {
				meta = m
			} else {
				meta = oldMap[d.OldLine]
			}
		}

		d.ResourceKind = meta.Kind
		d.ResourceAPIVersion = meta.APIVersion
		d.ResourceName = meta.Name
		d.ResourceNamespace = meta.Namespace
	}
}

func buildLineToResourceMetaMap(rendered string) map[int]kustomizeResourceMeta {
	lines := normalizeLines(rendered)
	out := make(map[int]kustomizeResourceMeta, len(lines))
	if len(lines) == 0 {
		return out
	}

	docStart := 0
	for i := 0; i <= len(lines); i++ {
		isBoundary := false
		if i == len(lines) {
			isBoundary = true
		} else if strings.TrimSpace(lines[i]) == "---" && i != docStart {
			isBoundary = true
		}
		if !isBoundary {
			continue
		}

		end := i - 1
		if end >= docStart {
			meta := parseResourceMetaFromDocLines(lines[docStart : end+1])
			for ln := docStart; ln <= end; ln++ {
				out[ln+1] = meta // 1-indexed
			}
		}
		docStart = i
	}

	return out
}

func normalizeLines(s string) []string {
	if s == "" {
		return []string{}
	}
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func parseResourceMetaFromDocLines(docLines []string) kustomizeResourceMeta {
	i := 0
	for i < len(docLines) {
		t := strings.TrimSpace(docLines[i])
		if t == "" || t == "---" {
			i++
			continue
		}
		break
	}
	if i >= len(docLines) {
		return kustomizeResourceMeta{}
	}

	content := strings.Join(docLines[i:], "\n")
	var obj struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Metadata   struct {
			Name      string `yaml:"name"`
			Namespace string `yaml:"namespace"`
		} `yaml:"metadata"`
	}
	if err := k8syaml.Unmarshal([]byte(content), &obj); err != nil {
		return kustomizeResourceMeta{}
	}
	return kustomizeResourceMeta{
		Kind:       obj.Kind,
		APIVersion: obj.APIVersion,
		Name:       obj.Metadata.Name,
		Namespace:  obj.Metadata.Namespace,
	}
}
