//revive:disable-next-line var-naming
package utils

import (
	"fmt"
	"strings"
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
	Type string
	Text string
	Line int
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

func DiffsToPrettyPrint(diffs []*ManifestDiff, verbose bool) string {
	var out strings.Builder
	outputDiffs := []*ManifestDiff{}

	if !verbose {
		for _, d := range diffs {
			if d.Type != "Equals" {
				outputDiffs = append(outputDiffs, d)
			}
		}
	} else {
		outputDiffs = diffs
	}

	for _, d := range outputDiffs {
		out.WriteString(fmt.Sprintf("%s%d %s %s\n", diffColorMap[d.Type], d.Line, diffSymbolMap[d.Type], d.Text))
	}

	return out.String()
}
