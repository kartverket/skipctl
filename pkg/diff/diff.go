package diff

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kartverket/skipctl/pkg/constants"
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
	Type string `json:"type"`
	Text string `json:"text"`
	Line int    `json:"line"`
}

type JSONOutput struct {
	File  string          `json:"file"`
	Ref   string          `json:"ref"`
	Diffs []*ManifestDiff `json:"diffs"`
}

func lcs(a, b []string) [][]int {
	// Returns a 2D table of LCS lengths
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := m - 1; i >= 0; i-- {
		for j := n - 1; j >= 0; j-- {
			switch {
			case a[i] == b[j]:
				dp[i][j] = dp[i+1][j+1] + 1
			case dp[i+1][j] >= dp[i][j+1]:
				dp[i][j] = dp[i+1][j]
			default:
				dp[i][j] = dp[i][j+1]
			}
		}
	}

	return dp
}

func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, "\n")
}

func LCS(a, b string) ([]*ManifestDiff, bool) {
	linesA := splitLines(a)
	linesB := splitLines(b)

	dp := lcs(linesA, linesB)
	i, j := 0, 0
	diffs := []*ManifestDiff{}
	hasDiff := false

	for i < len(linesA) && j < len(linesB) {
		switch {
		case linesA[i] == linesB[j]:
			diffs = append(diffs, &ManifestDiff{
				Type: constants.Equals,
				Text: linesA[i],
				Line: i + 1,
			})
			i++
			j++
		case dp[i+1][j] >= dp[i][j+1]:
			diffs = append(diffs, &ManifestDiff{
				Type: constants.Deletion,
				Text: linesA[i],
				Line: i + 1,
			})
			hasDiff = true
			i++
		default:
			diffs = append(diffs, &ManifestDiff{
				Type: constants.Insertion,
				Text: linesB[j],
				Line: j + 1,
			})
			hasDiff = true
			j++
		}
	}

	// Handle trailing insertions/deletions
	for i < len(linesA) {
		diffs = append(diffs, &ManifestDiff{
			Type: constants.Deletion,
			Text: linesA[i],
			Line: i,
		})
		hasDiff = true
		i++
	}
	for j < len(linesB) {
		diffs = append(diffs, &ManifestDiff{
			Type: constants.Insertion,
			Text: linesB[j],
			Line: j,
		})
		hasDiff = true
		j++
	}
	return diffs, hasDiff
}

func DiffsToPrettyPrint(diffs []*ManifestDiff, filename string) string {
	var out strings.Builder

	out.WriteString(fmt.Sprintf("%s%s%s\n", textBold, filename, colorReset))
	out.WriteString(formatPrettyHeader(diffs))

	for _, d := range diffs {
		out.WriteString(fmt.Sprintf("%s%d %s %s%s\n", diffColorMap[d.Type], d.Line, diffSymbolMap[d.Type], d.Text, colorReset))
	}
	return out.String()
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
	prevLineNum := -1
	hunkIdx := 0
	hunks := make(map[int][]*ManifestDiff)
	for _, diff := range diffs {
		if diff.Line > prevLineNum+1 {
			hunkIdx++
		}
		hunks[hunkIdx] = append(hunks[hunkIdx], diff)
		prevLineNum = diff.Line
	}
	for _, hunk := range hunks {
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
	var b strings.Builder
	startLineRemote, startLineLocal := 0, 0
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
	if startLineRemote == 0 {
		startLineRemote = startLineLocal - 1
	}
	var h strings.Builder
	fmt.Fprintf(&h, "@@ -%d,%d +%d,%d @@\n", startLineRemote, del+eql, startLineLocal, ins+eql)
	h.WriteString(b.String())
	return h.String()
}

func formatPrettyHeader(diffs []*ManifestDiff) string {
	startLineRemote, startLineLocal := 0, 0
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
	if startLineRemote == 0 {
		startLineRemote = startLineLocal - 1
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

func filterDiffsWithChunks(diffs []*ManifestDiff, nlines int) []*ManifestDiff {
	nonEqualDiffs := filterNonEqualDiffs(diffs)

	lineSet := make(map[int]struct{})
	for _, d := range nonEqualDiffs {
		for i := d.Line - nlines; i <= d.Line+nlines; i++ {
			if i > 0 {
				lineSet[i] = struct{}{}
			}
		}
	}
	outputDiffs := make([]*ManifestDiff, 0, len(diffs))
	for _, d := range diffs {
		if _, ok := lineSet[d.Line]; ok {
			outputDiffs = append(outputDiffs, d)
		}
	}
	return outputDiffs
}
