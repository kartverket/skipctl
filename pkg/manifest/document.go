package manifest

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/git"
	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/utils"
)

type Document struct {
	Name        string
	Extension   string
	Permissions os.FileMode
	Content     string
	Path        string
	FromStdin   bool
	Rendered    bool
}

func (d *Document) AtRef(ref string) (*Document, error) {
	content, err := git.GetFileContentAtRef(d.Name, ref)
	if err != nil {
		return nil, err
	}

	return &Document{
		Name:        d.Name,
		Extension:   d.Extension,
		Permissions: d.Permissions,
		Content:     *content,
		FromStdin:   false,
	}, nil
}

func (d *Document) Write(content string) error {
	d.Content = content
	if d.FromStdin {
		if _, err := io.WriteString(os.Stdout, d.Content); err != nil {
			return fmt.Errorf("error writing document to stdout: %w", err)
		}
		return nil
	}
	if err := os.WriteFile(d.Name, []byte(d.Content), d.Permissions); err != nil {
		return fmt.Errorf("error writing document to file: %w", err)
	}
	return nil
}

func FromFiles(filenames []string) ([]*Document, error) {
	var files = []*Document{}

	for _, filename := range filenames {
		fileContent, err := os.ReadFile(filename)

		if err != nil {
			logging.Logger().Error("unable to read file", "filename", filename)
			continue
		}
		finfo, ferr := os.Stat(filename)
		if ferr != nil {
			return nil, fmt.Errorf("unable to get file info %s %w", filename, ferr)
		}
		ext := strings.ToLower(filepath.Ext(filename))
		if filepath.Base(filename) == constants.ManifestKustomizeYaml || filepath.Base(filename) == constants.ManifestKustomizeYml {
			ext = constants.ManifestKustomizeYaml
		}
		files = append(files, &Document{
			Name:        filename,
			Extension:   ext,
			Permissions: finfo.Mode().Perm(),
			Content:     string(fileContent),
			FromStdin:   false,
		})
	}
	return files, nil
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
