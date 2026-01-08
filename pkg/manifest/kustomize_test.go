package manifest

import (
	"testing"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/diff"
)

func TestNormalizeLines_StripsTrailingEmptyLine(t *testing.T) {
	in := "a\nb\n"
	lines := normalizeLines(in)

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %#v", len(lines), lines)
	}
	if lines[0] != "a" || lines[1] != "b" {
		t.Fatalf("unexpected lines: %#v", lines)
	}

	empty := normalizeLines("")
	if len(empty) != 0 {
		t.Fatalf("expected 0 lines for empty input, got %d", len(empty))
	}
}

func TestParseResourceMetaFromDocLines_ParsesAndSkipsSeparators(t *testing.T) {
	doc := []string{
		"---",
		"",
		"apiVersion: apps/v1",
		"kind: Deployment",
		"metadata:",
		"  name: dep1",
		"  namespace: ns1",
		"spec:",
		"  replicas: 1",
	}

	meta := parseResourceMetaFromDocLines(doc)

	if meta.Kind != "Deployment" {
		t.Fatalf("expected Kind=Deployment, got %q", meta.Kind)
	}
	if meta.APIVersion != "apps/v1" {
		t.Fatalf("expected APIVersion=apps/v1, got %q", meta.APIVersion)
	}
	if meta.Name != "dep1" {
		t.Fatalf("expected Name=dep1, got %q", meta.Name)
	}
	if meta.Namespace != "ns1" {
		t.Fatalf("expected Namespace=ns1, got %q", meta.Namespace)
	}
}

func TestParseResourceMetaFromDocLines_ReturnsEmptyOnInvalidYAML(t *testing.T) {
	doc := []string{
		"apiVersion: v1",
		"kind: ConfigMap",
		"metadata:",
		"  name: [",
	}

	meta := parseResourceMetaFromDocLines(doc)
	if meta.Kind != "" || meta.APIVersion != "" || meta.Name != "" || meta.Namespace != "" {
		t.Fatalf("expected empty meta on invalid yaml, got %+v", meta)
	}
}

func TestBuildLineToResourceMetaMap_AssignsAcrossDocsAndSeparatorBelongsToNextDoc(t *testing.T) {
	rendered := "" +
		"apiVersion: v1\n" +
		"kind: ConfigMap\n" +
		"metadata:\n" +
		"  name: cm1\n" +
		"  namespace: ns1\n" +
		"data:\n" +
		"  a: b\n" +
		"---\n" +
		"apiVersion: apps/v1\n" +
		"kind: Deployment\n" +
		"metadata:\n" +
		"  name: dep1\n" +
		"  namespace: ns2\n" +
		"spec:\n" +
		"  replicas: 1\n"

	m := buildLineToResourceMetaMap(rendered)

	// the lines 1-7 should map to ConfigMap/cm1/ns1
	for ln := 1; ln <= 7; ln++ {
		meta, ok := m[ln]
		if !ok {
			t.Fatalf("expected metadata for line %d", ln)
		}
		if meta.Kind != "ConfigMap" || meta.APIVersion != "v1" || meta.Name != "cm1" || meta.Namespace != "ns1" {
			t.Fatalf("line %d: unexpected meta %+v", ln, meta)
		}
	}

	meta8, ok8 := m[8]
	if !ok8 {
		t.Fatalf("expected metadata for line 8")
	}
	if meta8.Kind != "Deployment" || meta8.APIVersion != "apps/v1" || meta8.Name != "dep1" || meta8.Namespace != "ns2" {
		t.Fatalf("line 8: unexpected meta %+v", meta8)
	}

	// the lines 9-15 should map to Deployment/dep1/ns2
	for ln := 9; ln <= 15; ln++ {
		meta, ok := m[ln]
		if !ok {
			t.Fatalf("expected metadata for line %d", ln)
		}
		if meta.Kind != "Deployment" || meta.APIVersion != "apps/v1" || meta.Name != "dep1" || meta.Namespace != "ns2" {
			t.Fatalf("line %d: unexpected meta %+v", ln, meta)
		}
	}
}

func TestAnnotateKustomizeDiffsWithResourceMeta_AssignsBasedOnDiffType(t *testing.T) {
	prevRendered := "" +
		"apiVersion: v1\n" +
		"kind: ConfigMap\n" +
		"metadata:\n" +
		"  name: cm1\n" +
		"  namespace: ns1\n" +
		"data:\n" +
		"  a: b\n" +
		"---\n" +
		"apiVersion: apps/v1\n" +
		"kind: Deployment\n" +
		"metadata:\n" +
		"  name: dep1\n" +
		"  namespace: ns2\n" +
		"spec:\n" +
		"  replicas: 1\n"

	rendered := prevRendered

	diffs := []*diff.ManifestDiff{
		{Type: constants.Deletion, Text: "  name: cm1", Line: 4, OldLine: 4, NewLine: 4},
		{Type: constants.Insertion, Text: "  name: dep1", Line: 12, OldLine: 12, NewLine: 12},
		{Type: constants.Equals, Text: "---", Line: 8, OldLine: 8, NewLine: 8},
		{Type: constants.Insertion, Text: "x", Line: 999, OldLine: 999, NewLine: 999},
	}

	annotateKustomizeDiffsWithResourceMeta(diffs, prevRendered, rendered)

	if diffs[0].ResourceKind != "ConfigMap" || diffs[0].ResourceAPIVersion != "v1" || diffs[0].ResourceName != "cm1" || diffs[0].ResourceNamespace != "ns1" {
		t.Fatalf("deletion: unexpected resource meta: %+v", diffs[0])
	}

	if diffs[1].ResourceKind != "Deployment" || diffs[1].ResourceAPIVersion != "apps/v1" || diffs[1].ResourceName != "dep1" || diffs[1].ResourceNamespace != "ns2" {
		t.Fatalf("insertion: unexpected resource meta: %+v", diffs[1])
	}

	if diffs[2].ResourceKind != "Deployment" || diffs[2].ResourceAPIVersion != "apps/v1" || diffs[2].ResourceName != "dep1" || diffs[2].ResourceNamespace != "ns2" {
		t.Fatalf("equals: unexpected resource meta: %+v", diffs[2])
	}

	if diffs[3].ResourceKind != "" || diffs[3].ResourceAPIVersion != "" || diffs[3].ResourceName != "" || diffs[3].ResourceNamespace != "" {
		t.Fatalf("out-of-range: expected empty resource meta, got: %+v", diffs[3])
	}
}
