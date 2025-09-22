//revive:disable-next-line var-naming
package utils

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
)

var diffSymbolMap = map[string]string{
	"Equals":    " ",
	"Insertion": "+",
	"Deletion":  "-",
}

var diffColorMap = map[string]string{
	"Equals":    colorReset,
	"Insertion": colorGreen,
	"Deletion":  colorRed,
}

type ManifestDiff struct {
	Type string `json:"type"`
	Text string `json:"text"`
	Line int    `json:"line"`
}

type DiffJSONOutput struct {
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

func DiffLCS(a, b string) ([]*ManifestDiff, bool) {
	linesA := strings.Split(a, "\n")
	linesB := strings.Split(b, "\n")
	dp := lcs(linesA, linesB)
	i, j := 0, 0
	diffs := []*ManifestDiff{}
	hasDiff := false

	for i < len(linesA) && j < len(linesB) {
		switch {
		case linesA[i] == linesB[j]:
			diffs = append(diffs, &ManifestDiff{
				Type: "Equals",
				Text: linesA[i],
				Line: i,
			})
			i++
			j++
		case dp[i+1][j] >= dp[i][j+1]:
			diffs = append(diffs, &ManifestDiff{
				Type: "Deletion",
				Text: linesA[i],
				Line: i,
			})
			hasDiff = true
			i++
		default:
			diffs = append(diffs, &ManifestDiff{
				Type: "Insertion",
				Text: linesB[j],
				Line: j,
			})
			hasDiff = true
			j++
		}
	}

	// Handle trailing insertions/deletions
	for i < len(linesA) {
		diffs = append(diffs, &ManifestDiff{
			Type: "Deletion",
			Text: linesA[i],
			Line: i,
		})
		hasDiff = true
		i++
	}
	for j < len(linesB) {
		diffs = append(diffs, &ManifestDiff{
			Type: "Insertion",
			Text: linesB[j],
			Line: j,
		})
		hasDiff = true
		j++
	}
	return diffs, hasDiff
}

func DiffsToPrettyPrint(diffs []*ManifestDiff) string {
	var out strings.Builder
	for _, d := range diffs {
		out.WriteString(fmt.Sprintf("%s%d %s %s%s\n", diffColorMap[d.Type], d.Line, diffSymbolMap[d.Type], d.Text, colorReset))
	}
	return out.String()
}

func DiffsToJSON(diffs []*ManifestDiff, fileName string, ref string) string {
	out := DiffJSONOutput{
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
		case "Insertion":
			insertions++
		case "Deletion":
			deletions++
		}
	}
	if insertions == 0 && deletions == 0 {
		return ""
	}

	// Header first, then diffs
	header := writePatchHeaders(fileName, insertions, deletions)
	body := writePatch(diffs)
	return header + body
}

func writePatchHeaders(fileName string, insertions, deletions int) string {
	var b strings.Builder
	// Shortstat
	fmt.Fprintf(&b, " %s | %d %s%s\n", fileName, insertions+deletions, strings.Repeat("+", insertions), strings.Repeat("-", deletions))
	fmt.Fprintf(&b, " 1 file changed, %d insertion(+), %d deletion(-)\n\n", insertions, deletions)

	// Patch headers
	fmt.Fprintf(&b, "diff --git a/%s b/%s\n", fileName, fileName)
	fmt.Fprintf(&b, "--- a/%s\n", fileName)
	fmt.Fprintf(&b, "+++ b/%s\n", fileName)
	return b.String()
}

func writePatch(diffs []*ManifestDiff) string {
	var b strings.Builder

	aLn, bLn := 1, 1
	i := 0
	for i < len(diffs) {
		if diffs[i].Type == "Equals" {
			aLn++
			bLn++
			i++
			continue
		}

		aStart, bStart := aLn, bLn
		aCount, bCount := 0, 0
		var h strings.Builder

		for i < len(diffs) && diffs[i].Type != "Equals" {
			switch diffs[i].Type {
			case "Deletion":
				h.WriteString("- ")
				h.WriteString(diffs[i].Text)
				h.WriteByte('\n')
				aCount++
				aLn++
			case "Insertion":
				h.WriteString("+ ")
				h.WriteString(diffs[i].Text)
				h.WriteByte('\n')
				bCount++
				bLn++
			}
			i++
		}

		// Emit hunk header and its lines.
		fmt.Fprintf(&b, "@@ -%d,%d +%d,%d @@\n", aStart, aCount, bStart, bCount)
		b.WriteString(h.String())
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
