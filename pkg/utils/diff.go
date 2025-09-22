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
	Type     string `json:"type"`
	Text     string `json:"text"`
	Line     int    `json:"line"`
	FileName string `json:"filename"`
	Ref      string `json:"ref"`
}

func Diff(a, b string) ([]*ManifestDiff, bool) {
	// Split input into lines
	linesA := strings.Split(a, "\n")
	linesB := strings.Split(b, "\n")

	maxLen := max(len(linesA), len(linesB))

	diffs := []*ManifestDiff{}
	hasDiff := false

	for i := range maxLen {
		var lineA, lineB string
		if i < len(linesA) {
			lineA = linesA[i]
		}
		if i < len(linesB) {
			lineB = linesB[i]
		}
		switch {
		case i >= len(linesA):
			diffs = append(diffs, &ManifestDiff{
				Type: "Insertion",
				Text: lineB,
				Line: i,
			})
			hasDiff = true
		case i >= len(linesB):
			diffs = append(diffs, &ManifestDiff{
				Type: "Deletion",
				Text: lineA,
				Line: i,
			})
			hasDiff = true
		case lineA != lineB:
			diffs = append(diffs, &ManifestDiff{
				Type: "Deletion",
				Text: lineA,
				Line: i,
			})

			diffs = append(diffs, &ManifestDiff{
				Type: "Insertion",
				Text: lineB,
				Line: i,
			})
			hasDiff = true
		default:
			diffs = append(diffs, &ManifestDiff{
				Type: "Equals",
				Text: lineA,
				Line: i,
			})
		}
	}

	return diffs, hasDiff
}

func filterDiffsVerbose(diffs []*ManifestDiff, verbose bool) []*ManifestDiff {
	if verbose {
		return diffs
	}

	outputDiffs := []*ManifestDiff{}
	for _, d := range diffs {
		if d.Type != constants.Equals {
			outputDiffs = append(outputDiffs, d)
		}
	}
	return outputDiffs
}

func DiffsToPrettyPrint(diffs []*ManifestDiff, verbose bool) string {
	filteredDiffs := filterDiffsVerbose(diffs, verbose)

	var out strings.Builder
	for _, d := range filteredDiffs {
		out.WriteString(fmt.Sprintf("%s%d %s %s\n", diffColorMap[d.Type], d.Line, diffSymbolMap[d.Type], d.Text))
	}
	return out.String()
}

func DiffsToJSON(diffs []*ManifestDiff, verbose bool) string {
	filteredDiffs := filterDiffsVerbose(diffs, verbose)

	// Marshal the filtered diffs to JSON
	jsonBytes, err := json.MarshalIndent(filteredDiffs, "", "  ")
	if err != nil {
		// Return an error string if marshalling fails
		return `{"error": "failed to marshal diffs to JSON"}`
	}
	return string(jsonBytes)
}

func DiffsToPatch(diffs []*ManifestDiff, verbose bool) string {
	filteredDiffs := filterDiffsVerbose(diffs, verbose)

	// return diff as patch format
	return makePatch(filteredDiffs)
}
func makePatch(diffs []*ManifestDiff) string {
	if len(diffs) == 0 {
		return ""
	}

	// Determine filename (fallback).
	var fileName string
	for _, d := range diffs {
		if fileName == "" && d.FileName != "" {
			fileName = d.FileName
		}
	}

	// Build stats for shortstat and unified header.
	var (
		insertions, deletions int
		aCount, bCount        int
		lineA, lineB          int
	)
	for _, d := range diffs {
		switch d.Type {
		case "Equals":
			aCount++
			bCount++
		case "Deletion":
			lineA = d.Line
			deletions++
			aCount++
		case "Insertion":
			lineB = d.Line
			insertions++
			bCount++
		}
	}
	if insertions == 0 && deletions == 0 {
		return ""
	}

	var b strings.Builder
	// Shortstat of the diff
	fmt.Fprintf(&b, " %s | %d %s%s\n", fileName, insertions+deletions, strings.Repeat("+", insertions), strings.Repeat("-", deletions))
	fmt.Fprintf(&b, " 1 file changed, %d insertions(+), %d deletions(-)\n\n",
		insertions, insertions)
	// Patch headers of the diff
	fmt.Fprintf(&b, "diff --git a /%s b /%s\n", fileName, fileName)
	fmt.Fprintf(&b, "--- a /%s\n", fileName)
	fmt.Fprintf(&b, "+++ b /%s\n", fileName)
	fmt.Fprintf(&b, "@@ -%d,%d +%d,%d @@\n", lineA, aCount, lineB, bCount)
	// Body of the diff
	for _, d := range diffs {
		switch d.Type {
		case "Equals":
			b.WriteByte(' ')
		case "Deletion":
			b.WriteByte('-')
		case "Insertion":
			b.WriteByte('+')
		default:
			b.WriteByte(' ')
		}
		b.WriteByte(' ')
		b.WriteString(d.Text)
		b.WriteByte('\n')
	}

	return b.String()
}
