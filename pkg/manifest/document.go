package manifest

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/utils"
)

type Document struct {
	Name        string
	Extension   string
	Permissions os.FileMode
	Content     string
	FromStdin   bool
}

func (r *Document) Write(content string) error {
	r.Content = content
	if r.FromStdin {
		if _, err := io.WriteString(os.Stdout, r.Content); err != nil {
			return fmt.Errorf("error writing document to stdout: %w", err)
		}
	}
	if err := os.WriteFile(r.Name, []byte(r.Content), r.Permissions); err != nil {
		return fmt.Errorf("error writing document to file: %w", err)
	}
	return nil
}

func NewDocuments(stdin bool, filenames []string) ([]*Document, error) {
	if stdin {
		return FromStdin()
	}
	return FromFiles(filenames), nil
}
func FromFiles(filenames []string) []*Document {
	var files = []*Document{}

	for _, filename := range filenames {
		fileContent, err := os.ReadFile(filename)

		if err != nil {
			logging.Logger().Error("unable to read file", "filename", filename)
			continue
		}
		files = append(files, &Document{
			Name:      filename,
			Extension: strings.ToLower(filepath.Ext(filename)),
			Content:   string(fileContent),
			FromStdin: false,
		})
	}
	return files
}
func FromStdin() ([]*Document, error) {
	log := logging.Logger()

	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}
	files := bytes.Split(content, []byte("\n---\n"))

	var manifestFiles []*Document

	for i, fileContent := range files {
		ext, detectFiletypeErr := utils.DetectFiletype(fileContent)
		if detectFiletypeErr != nil {
			log.Error(detectFiletypeErr.Error(), "reason", "skipping")
			continue
		}
		manifestFiles = append(manifestFiles, &Document{
			Name:      fmt.Sprintf("stdin_%d", i),
			Extension: ext,
			Content:   string(fileContent),
			FromStdin: true,
		})
	}
	return manifestFiles, nil
}
