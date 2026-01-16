package diff

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/sergi/go-diff/diffmatchpatch"
)

const (
	colorRed   = "\x1b[31m"
	colorGreen = "\x1b[32m"
	colorReset = "\x1b[0m"
	textBold   = "\033[1m"
)

var diffSymbolMap = map[string]string{
	constants.Equals:    " ",
	constants.Insertion: "+",
	constants.Deletion:  "-",
}

var diffColorMap = map[string]string{
	constants.Equals:    colorReset,
	constants.Insertion: colorGreen,
	constants.Deletion:  colorRed,
}

type ManifestDiff struct {
	Type    string `json:"type"`
	Text    string `json:"text"`
	Line    int    `json:"line"`
	OldLine int    `json:"old_line"`
	NewLine int    `json:"new_line"`

	ResourceKind       string `json:"resource_kind,omitempty"`
	ResourceName       string `json:"resource_name,omitempty"`
	ResourceNamespace  string `json:"resource_namespace,omitempty"`
	ResourceAPIVersion string `json:"resource_api_version,omitempty"`
}

type JSONOutput struct {
	File  string          `json:"file"`
	Ref   string          `json:"ref"`
	Diffs []*ManifestDiff `json:"diffs"`
}

func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, "\n")
}
func convertToManifestDiff(dmpDiffs []diffmatchpatch.Diff) ([]*ManifestDiff, bool) {
	diffs := make([]*ManifestDiff, 0)
	hasDiff := false
	lineA := 1
	lineB := 1

	for _, d := range dmpDiffs {
		lines := splitLines(d.Text)
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			diffs, lineA, lineB = processEqualLines(diffs, lines, lineA, lineB)
		case diffmatchpatch.DiffDelete:
			diffs, lineA = processDeletionLines(diffs, lines, lineA, lineB)
			hasDiff = true
		case diffmatchpatch.DiffInsert:
			diffs, lineB = processInsertionLines(diffs, lines, lineB, lineA)
			hasDiff = true
		}
	}
	return diffs, hasDiff
}

func processEqualLines(diffs []*ManifestDiff, lines []string, lineA, lineB int) ([]*ManifestDiff, int, int) {
	for i, line := range lines {
		if shouldSkipLine(i, len(lines), line) {
			continue
		}
		diffs = append(diffs, &ManifestDiff{
			Type:    constants.Equals,
			Text:    line,
			Line:    lineA,
			OldLine: lineA,
			NewLine: lineB,
		})
		lineA++
		lineB++
	}
	return diffs, lineA, lineB
}

func processDeletionLines(diffs []*ManifestDiff, lines []string, lineA, lineB int) ([]*ManifestDiff, int) {
	for i, line := range lines {
		if shouldSkipLine(i, len(lines), line) {
			continue
		}
		diffs = append(diffs, &ManifestDiff{
			Type:    constants.Deletion,
			Text:    line,
			Line:    lineA,
			OldLine: lineA,
			NewLine: lineB,
		})
		lineA++
	}
	return diffs, lineA
}

func processInsertionLines(diffs []*ManifestDiff, lines []string, lineB, lineA int) ([]*ManifestDiff, int) {
	for i, line := range lines {
		if shouldSkipLine(i, len(lines), line) {
			continue
		}
		diffs = append(diffs, &ManifestDiff{
			Type:    constants.Insertion,
			Text:    line,
			Line:    lineB,
			OldLine: lineA,
			NewLine: lineB,
		})
		lineB++
	}
	return diffs, lineB
}

func shouldSkipLine(index, totalLines int, line string) bool {
	// Skip the last empty line if text ended with newline
	return index == totalLines-1 && line == ""
}
func CalculateDiff(a, b string) ([]*ManifestDiff, bool) {
	dmp := diffmatchpatch.New()
	// Convert to runes for fast diff algorithm with DiffMainRunes
	textA, textB, lineArray := dmp.DiffLinesToRunes(a, b)
	dmpDiffs := dmp.DiffMainRunes(textA, textB, true)

	// Convert back from encoded strings to lines
	dmpDiffs = dmp.DiffCharsToLines(dmpDiffs, lineArray)
	return convertToManifestDiff(dmpDiffs)
}

func DiffsToPrettyPrint(diffs []*ManifestDiff, filename string) string {
	var out strings.Builder

	if len(diffs) == 0 {
		return ""
	}

	out.WriteString(fmt.Sprintf("%s%s%s\n", textBold, filename, colorReset))
	hunks := groupIntoHunks(diffs)

	for i, hunk := range hunks {
		if len(hunk) == 0 {
			continue
		}
		if i > 0 {
			out.WriteByte('\n')
		}
		if hdr := resourceHeaderForHunk(hunk); hdr != "" {
			out.WriteString(fmt.Sprintf("%s%s%s\n", textBold, hdr, colorReset))
		}
		out.WriteString(formatPrettyHeader(hunk))
		for _, d := range hunk {
			out.WriteString(fmt.Sprintf("%s%d %s %s%s\n",
				diffColorMap[d.Type], d.Line, diffSymbolMap[d.Type], d.Text, colorReset))
		}
	}
	return out.String()
}

func resourceHeaderForHunk(hunk []*ManifestDiff) string {
	d := pickResourceDiff(hunk)
	if d == nil {
		return ""
	}
	return formatResourceHeader(d.ResourceKind, d.ResourceAPIVersion, d.ResourceNamespace, d.ResourceName)
}

func pickResourceDiff(hunk []*ManifestDiff) *ManifestDiff {
	// choose resource from changed lines first
	for _, d := range hunk {
		if d.Type != constants.Equals && diffHasResourceMeta(d) {
			return d
		}
	}
	// if for some reason none found, choose any with resource information
	for _, d := range hunk {
		if diffHasResourceMeta(d) {
			return d
		}
	}
	return nil
}

func formatResourceHeader(kind, apiVersion, namespace, name string) string {
	if kind == "" && name == "" {
		return ""
	}
	left := "Resource"
	if kind != "" {
		left = kind
		if apiVersion != "" {
			left = left + "." + apiVersion
		}
	}
	if name == "" {
		return left
	}
	if namespace != "" {
		return left + "/" + namespace + "/" + name
	}
	return left + "/" + name
}

func diffHasResourceMeta(d *ManifestDiff) bool {
	return d.ResourceKind != "" || d.ResourceName != "" || d.ResourceNamespace != "" || d.ResourceAPIVersion != ""
}

// groups diffs into diff hunks based on continuity between oldLine and newLine.
func groupIntoHunks(entries []*ManifestDiff) [][]*ManifestDiff {
	if len(entries) == 0 {
		return nil
	}

	var hunks [][]*ManifestDiff
	var currentHunk []*ManifestDiff

	expectedOldLine, expectedNewLine := -1, -1

	for i, entry := range entries {
		startNewHunk := i == 0 || entry.OldLine != expectedOldLine || entry.NewLine != expectedNewLine
		if startNewHunk {
			if len(currentHunk) > 0 {
				hunks = append(hunks, currentHunk)
			}
			currentHunk = nil
		}

		currentHunk = append(currentHunk, entry)
		expectedOldLine, expectedNewLine = nextExpectedLines(entry)
	}

	if len(currentHunk) > 0 {
		hunks = append(hunks, currentHunk)
	}

	return hunks
}

func nextExpectedLines(d *ManifestDiff) (int, int) {
	switch d.Type {
	case constants.Equals:
		return d.OldLine + 1, d.NewLine + 1
	case constants.Deletion:
		return d.OldLine + 1, d.NewLine
	case constants.Insertion:
		return d.OldLine, d.NewLine + 1
	default:
		return d.OldLine, d.NewLine
	}
}

func DiffsToJSON(diffs []*ManifestDiff, fileName string, ref string) string {
	out := JSONOutput{
		File:  fileName,
		Ref:   ref,
		Diffs: diffs,
	}
	jsonBytes, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return `{"error": "failed to marshal diffs to JSON"}`
	}
	return string(jsonBytes)
}

func DiffsToPatch(diffs []*ManifestDiff, fileName string) string {
	if len(diffs) == 0 {
		return ""
	}

	// Count insertions/deletions for shortstat.
	insertions, deletions := 0, 0
	for _, d := range diffs {
		switch d.Type {
		case constants.Insertion:
			insertions++
		case constants.Deletion:
			deletions++
		}
	}
	if insertions == 0 && deletions == 0 {
		return ""
	}

	// Header first, then diffs
	header := writePatchHeaders(fileName)
	body := writePatch(diffs)
	return header + body
}

func writePatchHeaders(fileName string) string {
	var b strings.Builder
	// Shortstat
	fmt.Fprintf(&b, "--- remote /%s\n", fileName)
	fmt.Fprintf(&b, "+++ local /%s\n", fileName)
	return b.String()
}

func writePatch(diffs []*ManifestDiff) string {
	var b strings.Builder
	for _, hunk := range groupIntoHunks(diffs) {
		b.WriteString(formatPatchHunk(hunk))
	}
	return b.String()
}

func FilterDiffs(diffs []*ManifestDiff, verbosityLevel string, chunkSize int) []*ManifestDiff {
	switch verbosityLevel {
	case constants.DiffVerbosityFull:
		return diffs
	case constants.DiffVerbosityChunk:
		return filterDiffsWithChunks(diffs, chunkSize)
	case constants.DiffVerbosityMinimal:
		return filterNonEqualDiffs(diffs)
	default:
		return diffs
	}
}

func formatPatchHunk(hunk []*ManifestDiff) string {
	if len(hunk) == 0 {
		return ""
	}
	var b strings.Builder
	startLineRemote, startLineLocal := hunk[0].OldLine, hunk[0].NewLine
	ins, del, eql := 0, 0, 0
	for _, h := range hunk {
		switch h.Type {
		case constants.Deletion:
			if startLineRemote == 0 {
				startLineRemote = h.Line
			}
			b.WriteString("- ")
			b.WriteString(h.Text)
			b.WriteByte('\n')
			del++
		case constants.Insertion:
			if startLineLocal == 0 {
				startLineLocal = h.Line
			}
			b.WriteString("+ ")
			b.WriteString(h.Text)
			b.WriteByte('\n')
			ins++
		case constants.Equals:
			b.WriteString("  ")
			b.WriteString(h.Text)
			b.WriteByte('\n')
			eql++
		}
	}
	if del+eql == 0 {
		startLineRemote--
	}
	if ins+eql == 0 {
		startLineLocal--
	}
	var h strings.Builder
	fmt.Fprintf(&h, "@@ -%d,%d +%d,%d @@\n", startLineRemote, del+eql, startLineLocal, ins+eql)
	h.WriteString(b.String())
	return h.String()
}

func formatPrettyHeader(diffs []*ManifestDiff) string {
	startLineRemote, startLineLocal := diffs[0].OldLine, diffs[0].NewLine
	ins, del, eql := 0, 0, 0
	for _, h := range diffs {
		switch h.Type {
		case constants.Deletion:
			if startLineRemote == 0 {
				startLineRemote = h.Line
			}
			del++
		case constants.Insertion:
			if startLineLocal == 0 {
				startLineLocal = h.Line
			}
			ins++
		case constants.Equals:
			eql++
		}
	}
	if del+eql == 0 {
		startLineRemote--
	}
	if ins+eql == 0 {
		startLineLocal--
	}

	return fmt.Sprintf("%s@@ %s-%d,%d %s+%d,%d @@%s\n", textBold, colorRed, startLineRemote, del+eql, colorGreen, startLineLocal, ins+eql, colorReset)
}

func filterNonEqualDiffs(diffs []*ManifestDiff) []*ManifestDiff {
	outputDiffs := make([]*ManifestDiff, 0, len(diffs))
	for _, d := range diffs {
		if d.Type != constants.Equals {
			outputDiffs = append(outputDiffs, d)
		}
	}
	return outputDiffs
}
func filterNonEqualDiffsWithIndices(diffs []*ManifestDiff) map[int]*ManifestDiff {
	outDiffsWithIdx := make(map[int]*ManifestDiff)
	for idx, diff := range diffs {
		if diff.Type != constants.Equals {
			outDiffsWithIdx[idx] = diff
		}
	}
	return outDiffsWithIdx
}

func filterDiffsWithChunks(diffs []*ManifestDiff, nlines int) []*ManifestDiff {
	// Filter out diffs and the changed indices
	nonEqualDiffsWithIdx := filterNonEqualDiffsWithIndices(diffs)
	if nonEqualDiffsWithIdx == nil {
		return []*ManifestDiff{}
	}
	// Indices to keep for correct lines of context
	ctxIndices := make(map[int]struct{}) // Want the map, but not the data therefor empty struct
	for idx := range nonEqualDiffsWithIdx {
		ctxIndices[idx] = struct{}{}
		// Keep context lines before
		for i := 1; i <= nlines; i++ {
			if idx-i >= 0 {
				ctxIndices[idx-i] = struct{}{}
			}
			// Keep context lines after
			if idx+i < len(diffs) {
				ctxIndices[idx+i] = struct{}{}
			}
		}
	}
	outputDiffs := make([]*ManifestDiff, 0, len(diffs))
	for i, diff := range diffs {
		if _, ok := ctxIndices[i]; ok {
			outputDiffs = append(outputDiffs, diff)
		}
	}
	return outputDiffs
}
