//nolint:revive // the package name is intentional
package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func FindManifestFiles(path string) ([]string, error) {
	files, err := FindFilesWithSuffixes(path, constants.ManifestSuffixes)

	if err != nil {
		return []string{}, fmt.Errorf("error collecting files: %v", err.Error())
	}

	if len(files) == 0 {
		return []string{}, fmt.Errorf("no manifests found in path %s", path)
	}

	return files, nil
}

func FindFilesWithSuffixes(directory string, suffixes []string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(directory, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))

			if slices.Contains(suffixes, ext) {
				files = append(files, path)
			}
		}
		return nil
	})
	return files, err
}

// CreateTempDirectory creates a temporary directory with a given name in a unique location.
func CreateTempDirectory(name string) (string, error) {
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("skipctl-%s-*", name))

	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	return tempDir, nil
}

// CopyFilesToDirectory copies files from an filesystem to a specified local directory.
//
// It walks through the filesystem and writes each file to the destination directory,
// preserving the directory structure. Directories are created as needed.
func CopyFilesToDirectory(filesystem fs.FS, destinationDir string) error {
	return fs.WalkDir(filesystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		src, openErr := filesystem.Open(path)
		if openErr != nil {
			return openErr
		}
		defer src.Close()
		destPath := filepath.Join(destinationDir, path)
		// Create the directory if it doesn't exist
		if err = os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return err
		}
		// Create/truncate the destination file.
		dst, createErr := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
		if createErr != nil {
			return createErr
		}
		_, copyErr := io.Copy(dst, src)
		closeErr := dst.Close() // close explicitly (don’t defer inside loop)
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
}

func DetectFiletype(content []byte) (string, error) {
	trimmed := bytes.TrimLeft(content, " \t\r\n")
	if len(trimmed) == 0 {
		return "", errors.New("unable to detect filetype: Empty file content")
	}

	jsonnetPrefixes := [][]byte{
		[]byte("local "),
		[]byte("import "),
		[]byte("importstr "),
		[]byte("importbin "),
		[]byte("function"),
	}
	for _, p := range jsonnetPrefixes {
		if bytes.HasPrefix(trimmed, p) {
			return constants.ManifestSuffixJsonnet, nil
		}
	}

	firstChar := trimmed[0]

	switch firstChar {
	case '{', '[':
		return constants.ManifestSuffixJsonnet, nil
	default:
		return constants.ManifestSuffixYaml, nil
	}
}

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
