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
func NewKustomizeRenderer(output *slog.Logger, sortOutput bool, dividerEnabled ...bool) *KustomizeRenderer {
	opts := krusty.MakeDefaultOptions()

	// SKIP kustomize options
	opts.PluginConfig.HelmConfig.Enabled = true
	opts.PluginConfig.HelmConfig.Command = "helm"
	opts.LoadRestrictions = types.LoadRestrictionsNone

	if sortOutput {
		opts.Reorder = "legacy"
	}

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
func NewKustomizeDiffer(source Source, sortOutput bool) *KustomizeDiffer {
	currentBuf := &bytes.Buffer{}
	prevBuf := &bytes.Buffer{}

	currentLogger := logging.NewRawLoggerTo(currentBuf)
	prevLogger := logging.NewRawLoggerTo(prevBuf)

	return &KustomizeDiffer{
		source:          source,
		currentBuffer:   currentBuf,
		prevBuffer:      prevBuf,
		currentRenderer: NewKustomizeRenderer(currentLogger, sortOutput),
		prevRenderer:    NewKustomizeRenderer(prevLogger, sortOutput),
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
		meta := metaForDiffLine(d, oldMap, newMap)

		d.ResourceKind = meta.Kind
		d.ResourceAPIVersion = meta.APIVersion
		d.ResourceName = meta.Name
		d.ResourceNamespace = meta.Namespace
	}
}

func metaForDiffLine(d *diff.ManifestDiff, oldMap, newMap map[int]kustomizeResourceMeta) kustomizeResourceMeta {
	switch d.Type {
	case constants.Insertion:
		return newMap[d.NewLine]
	case constants.Deletion:
		return oldMap[d.OldLine]
	default:
		// if there is change in metadata between new and old, prefer new.
		if m, ok := newMap[d.NewLine]; ok {
			return m
		}
		return oldMap[d.OldLine]
	}
}

func buildLineToResourceMetaMap(rendered string) map[int]kustomizeResourceMeta {
	lines := normalizeLines(rendered)
	out := make(map[int]kustomizeResourceMeta, len(lines))
	if len(lines) == 0 {
		return out
	}

	for _, r := range splitYAMLDocRanges(lines) {
		meta := parseResourceMetaFromDocLines(lines[r.start : r.end+1])
		for i := r.start; i <= r.end; i++ {
			out[i+1] = meta // 1-indexed
		}
	}
	return out
}

type docRange struct{ start, end int }

func splitYAMLDocRanges(lines []string) []docRange {
	starts := []int{0}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			starts = append(starts, i)
		}
	}
	ranges := make([]docRange, 0, len(starts))
	for idx, start := range starts {
		end := len(lines) - 1
		if idx+1 < len(starts) {
			end = starts[idx+1] - 1
		}
		if end >= start {
			ranges = append(ranges, docRange{start: start, end: end})
		}
	}
	return ranges
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
