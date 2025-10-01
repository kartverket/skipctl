package diff

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/kartverket/skipctl/pkg/constants"
)

func TestLCS_NoDiff(t *testing.T) {
	a := "line1\nline2\nline3"
	b := "line1\nline2\nline3"

	diffs, hasDiff := LCS(a, b)

	if hasDiff {
		t.Fatalf("expected hasDiff=false, got true")
	}
	if len(diffs) != 3 {
		t.Fatalf("expected 3 diffs, got %d", len(diffs))
	}
	for i, d := range diffs {
		if d.Type != constants.Equals {
			t.Fatalf("diff %d expected Equals, got %s", i, d.Type)
		}
		if d.Line != i+1 {
			t.Fatalf("diff %d expected Line=%d, got %d", i, i+1, d.Line)
		}
	}
}

func TestLCS_InsertDelete(t *testing.T) {
	a := "a1\na2\na3"
	b := "a1\nb2\na3\nb4"

	diffs, hasDiff := LCS(a, b)
	if !hasDiff {
		t.Fatalf("expected hasDiff=true, got false")
	}

	// Expected sequence:
	// = a1 (line 1)
	// - a2 (line 2 in A)
	// + b2 (line 2 in B)
	// = a3 (line 3)
	// + b4 (trailing insertion, line index is j during trailing phase which code sets to j)
	gotTypes := make([]string, 0, len(diffs))
	gotTexts := make([]string, 0, len(diffs))
	for _, d := range diffs {
		gotTypes = append(gotTypes, d.Type)
		gotTexts = append(gotTexts, d.Text)
	}
	expectedTypes := []string{
		constants.Equals,
		constants.Deletion,
		constants.Insertion,
		constants.Equals,
		constants.Insertion,
	}
	expectedTexts := []string{"a1", "a2", "b2", "a3", "b4"}

	if strings.Join(gotTypes, ",") != strings.Join(expectedTypes, ",") {
		t.Fatalf("unexpected type sequence: got %v want %v", gotTypes, expectedTypes)
	}
	if strings.Join(gotTexts, ",") != strings.Join(expectedTexts, ",") {
		t.Fatalf("unexpected text sequence: got %v want %v", gotTexts, expectedTexts)
	}
}

func TestDiffsToPrettyPrint_Basic(t *testing.T) {
	diffs := []*ManifestDiff{
		{Type: constants.Equals, Text: "same", Line: 1},
		{Type: constants.Insertion, Text: "add", Line: 2},
		{Type: constants.Deletion, Text: "del", Line: 3},
	}

	out := DiffsToPrettyPrint(diffs)

	// Contains line numbers, symbols, and texts in order
	wantLines := []string{
		"1   same", // space symbol for equals
		"2 + add",
		"3 - del",
	}
	for _, w := range wantLines {
		if !strings.Contains(out, w) {
			t.Fatalf("pretty output missing %q in:\n%s", w, out)
		}
	}
	// Ensure color reset sequences exist at least once
	if !strings.Contains(out, "\x1b[0m") {
		t.Fatalf("expected color reset in output")
	}
}

func TestDiffsToJSON_Structure(t *testing.T) {
	diffs := []*ManifestDiff{
		{Type: constants.Equals, Text: "same", Line: 1},
		{Type: constants.Insertion, Text: "add", Line: 2},
	}
	file := "file.txt"
	ref := "main@abc123"

	js := DiffsToJSON(diffs, file, ref)

	var parsed struct {
		File  string         `json:"file"`
		Ref   string         `json:"ref"`
		Diffs []ManifestDiff `json:"diffs"`
	}
	if err := json.Unmarshal([]byte(js), &parsed); err != nil {
		t.Fatalf("failed to unmarshal json: %v\n%s", err, js)
	}
	if parsed.File != file || parsed.Ref != ref {
		t.Fatalf("unexpected header fields: got file=%q ref=%q", parsed.File, parsed.Ref)
	}
	if len(parsed.Diffs) != 2 {
		t.Fatalf("expected 2 diffs, got %d", len(parsed.Diffs))
	}
	if parsed.Diffs[1].Type != constants.Insertion || parsed.Diffs[1].Text != "add" || parsed.Diffs[1].Line != 2 {
		t.Fatalf("unexpected second diff: %+v", parsed.Diffs[1])
	}
}

func TestDiffsToPatch_EmptyOrNoChanges(t *testing.T) {
	// Empty diffs -> empty patch
	out := DiffsToPatch(nil, "x.txt")
	if out != "" {
		t.Fatalf("expected empty patch for nil diffs, got:\n%s", out)
	}

	// Only equals -> empty patch
	eq := []*ManifestDiff{
		{Type: constants.Equals, Text: "a", Line: 1},
		{Type: constants.Equals, Text: "b", Line: 2},
	}
	out = DiffsToPatch(eq, "x.txt")
	if out != "" {
		t.Fatalf("expected empty patch for equals-only diffs, got:\n%s", out)
	}
}

func TestDiffsToPatch_WithChangesAndHunks(t *testing.T) {
	// Two separated change regions to exercise hunk splitting
	diffs := []*ManifestDiff{
		{Type: constants.Equals, Text: "a1", Line: 1},
		{Type: constants.Deletion, Text: "a2", Line: 2},
		{Type: constants.Insertion, Text: "b2", Line: 2},
		{Type: constants.Equals, Text: "a3", Line: 3},
		// gap here (simulate jump)
		{Type: constants.Equals, Text: "a10", Line: 10},
		{Type: constants.Insertion, Text: "b11", Line: 11},
	}

	out := DiffsToPatch(diffs, "file.txt")

	// Header lines present
	if !strings.HasPrefix(out, "--- remote /file.txt\n+++ local /file.txt\n") {
		t.Fatalf("missing or wrong headers:\n%s", out)
	}

	// Should contain two hunks with @@ headers
	hunkCount := strings.Count(out, "@@ ")
	if hunkCount != 2 {
		t.Fatalf("expected 2 hunks, got %d\n%s", hunkCount, out)
	}

	// Check that change lines include +/- and equals lines include two spaces
	if !strings.Contains(out, "- a2\n") {
		t.Fatalf("missing deletion line:\n%s", out)
	}
	if !strings.Contains(out, "+ b2\n") {
		t.Fatalf("missing insertion line:\n%s", out)
	}
	if !strings.Contains(out, "  a3\n") {
		t.Fatalf("missing context equals line:\n%s", out)
	}
	if !strings.Contains(out, "+ b11\n") {
		t.Fatalf("missing second hunk insertion line:\n%s", out)
	}
}

func TestFilterNonEqualDiffs(t *testing.T) {
	diffs := []*ManifestDiff{
		{Type: constants.Equals, Text: "x", Line: 1},
		{Type: constants.Insertion, Text: "y", Line: 2},
		{Type: constants.Deletion, Text: "z", Line: 3},
		{Type: constants.Equals, Text: "w", Line: 4},
	}
	filtered := callFilterNonEqualDiffs(t, diffs)

	if len(filtered) != 2 {
		t.Fatalf("expected 2 diffs, got %d", len(filtered))
	}
	if filtered[0].Type != constants.Insertion || filtered[0].Text != "y" {
		t.Fatalf("unexpected first filtered diff: %+v", filtered[0])
	}
	if filtered[1].Type != constants.Deletion || filtered[1].Text != "z" {
		t.Fatalf("unexpected second filtered diff: %+v", filtered[1])
	}
}

// helper to reach unexported by using exported FilterDiffs with Minimal mode equivalence.
func callFilterNonEqualDiffs(t *testing.T, diffs []*ManifestDiff) []*ManifestDiff {
	t.Helper()
	got := FilterDiffs(diffs, constants.DiffVerbosityMinimal, 0)
	return got
}

func TestFilterDiffsWithChunks(t *testing.T) {
	// Build a sequence with equals around a change at line 5
	diffs := []*ManifestDiff{
		{Type: constants.Equals, Text: "1", Line: 1},
		{Type: constants.Equals, Text: "2", Line: 2},
		{Type: constants.Equals, Text: "3", Line: 3},
		{Type: constants.Equals, Text: "4", Line: 4},
		{Type: constants.Insertion, Text: "5+", Line: 5},
		{Type: constants.Equals, Text: "6", Line: 6},
		{Type: constants.Equals, Text: "7", Line: 7},
		{Type: constants.Equals, Text: "8", Line: 8},
	}
	// Chunk size = 1 should keep lines 4..6
	got := FilterDiffs(diffs, constants.DiffVerbosityChunk, 1)

	keptLines := make([]int, 0, len(got))
	for _, d := range got {
		keptLines = append(keptLines, d.Line)
	}
	want := []int{4, 5, 6}
	if strings.TrimSpace(intSliceToString(keptLines)) != strings.TrimSpace(intSliceToString(want)) {
		t.Fatalf("unexpected kept lines: got %v want %v", keptLines, want)
	}
}

func intSliceToString(v []int) string {
	return strings.Trim(fmt.Sprint(v), "[]")
}
