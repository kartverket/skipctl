package utils

import (
	"fmt"
	"io"
	"io/fs"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kartverket/skipctl/pkg/constants"
	"go.yaml.in/yaml/v4"
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
	"github.com/kartverket/skipctl/pkg/logging"
	"go.yaml.in/yaml/v4"
)

type ManifestFile struct {
	Name      string
	Extension string
	Content   string
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

func UnmarshalYamlFromFile(filename string) (any, error) {
	fileContents, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var output any

	err = yaml.Unmarshal(fileContents, &output)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func MarshalYamlFromFile(filename string) ([]byte, error) {
	fileContents, err := UnmarshalYamlFromFile(filename)
	if err != nil {
		return nil, err
	}

	marshalled, err := yaml.Marshal(fileContents)
	if err != nil {
		return nil, err
	}
	return marshalled, nil
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
func ReadFiles(filenames []string) []*ManifestFile {
	var files = []*ManifestFile{}

	for _, filename := range filenames {
		fileContent, err := os.ReadFile(filename)

		if err != nil {
			logging.Logger().Error("unable to read file", "filename", filename)
			continue
		}
		files = append(files, &ManifestFile{
			Name:      filename,
			Extension: strings.ToLower(filepath.Ext(filename)),
			Content:   string(fileContent),
		})
	}
	return files
}

func detectFiletype(content []byte) (string, error) {
	trimmed := bytes.TrimLeft(content, " \t\r\n")
	if len(trimmed) == 0 {
		return "", fmt.Errorf("Unable to detect filetype: Empty file content")
	}
	firstChar := trimmed[0]

	switch firstChar {
	case '{', '[':
		return constants.ManifestSuffixJsonnet, nil
	default:
		return constants.ManifestSuffixYaml, nil
	}
}

func ReadFilesFromStdin() ([]*ManifestFile, error) {
	log := logging.Logger()

	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}
	files := bytes.Split(content, []byte("\n---\n"))

	var manifestFiles []*ManifestFile

	for i, fileContent := range files {
		ext, err := detectFiletype(fileContent)
		if err != nil {
			log.Error(err.Error(), "reason", "skipping")
			continue
		}
		manifestFiles = append(manifestFiles, &ManifestFile{
			Name:      fmt.Sprintf("stdin_%d", i),
			Extension: ext,
			Content:   string(fileContent),
		})

	}
	return manifestFiles, nil
}
