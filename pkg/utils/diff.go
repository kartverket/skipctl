package utils

import (
	"fmt"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

const (
	colorRed   = "\x1b[31m"
	colorGreen = "\x1b[32m"
	colorReset = "\x1b[0m"
)

func Diff(a, b string, verbose bool) (string, bool) {
	dmp := diffmatchpatch.New()

	ar, br, lineArray := dmp.DiffLinesToRunes(a, b)
	diffs := dmp.DiffMainRunes(ar, br, false)
	diffs = dmp.DiffCleanupSemantic(diffs)
	diffs = dmp.DiffCleanupEfficiency(diffs)
	diffs = dmp.DiffCharsToLines(diffs, lineArray)

	allEqual := true
	var out strings.Builder
	lineA, lineB := 1, 1

	for _, d := range diffs {
		handleDiffChunk(&out, d, verbose, &lineA, &lineB)
		if d.Type.String() != "Equal" {
			allEqual = false
		}
	}

	return out.String(), !allEqual
}

func handleDiffChunk(out *strings.Builder, d diffmatchpatch.Diff, verbose bool, lineA, lineB *int) {
	lines := splitKeepNL(d.Text)
	switch d.Type {
	case diffmatchpatch.DiffEqual:
		writeEqual(out, lines, verbose, lineA, lineB)
	case diffmatchpatch.DiffDelete:
		writeDelete(out, lines, lineA)
	case diffmatchpatch.DiffInsert:
		writeInsert(out, lines, lineB)
	}
}

func writeEqual(out *strings.Builder, lines []string, verbose bool, lineA, lineB *int) {
	if verbose {
		cur := *lineA
		for _, ln := range lines {
			fmt.Fprintf(out, "%d  ", cur)
			out.WriteString(ln)
			if strings.HasSuffix(ln, "\n") {
				cur++
			}
		}
		*lineA = cur
		*lineB = cur
		return
	}
	inc := countNewlines(lines)
	*lineA += inc
	*lineB += inc
}

func writeDelete(out *strings.Builder, lines []string, lineA *int) {
	cur := *lineA
	for _, ln := range lines {
		out.WriteString(colorRed)
		fmt.Fprintf(out, "%d -", cur)
		out.WriteString(ln)
		out.WriteString(colorReset)
		if strings.HasSuffix(ln, "\n") {
			cur++
		}
	}
	*lineA = cur
}

func writeInsert(out *strings.Builder, lines []string, lineB *int) {
	cur := *lineB
	for _, ln := range lines {
		out.WriteString(colorGreen)
		fmt.Fprintf(out, "%d +", cur)
		out.WriteString(ln)
		out.WriteString(colorReset)
		if strings.HasSuffix(ln, "\n") {
			cur++
		}
	}
	*lineB = cur
}

func countNewlines(lines []string) int {
	n := 0
	for _, ln := range lines {
		if strings.HasSuffix(ln, "\n") {
			n++
		}
	}
	return n
}

// splitKeepNL splits on '\n' and keeps newline characters at the end of each chunk.
func splitKeepNL(s string) []string {
	if s == "" {
		return nil
	}
	var res []string
	start := 0
	for i := range s {
		if s[i] == '\n' {
			res = append(res, s[start:i+1])
			start = i + 1
		}
	}
	if start < len(s) {
		res = append(res, s[start:])
	}
	return res
}
