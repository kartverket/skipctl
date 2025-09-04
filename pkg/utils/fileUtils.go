package utils

import (
	"embed"
	"fmt"
	"io/fs"
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

func CreateTempDirectory(pattern string) (string, error) {
	tempDir, err := os.MkdirTemp("", pattern)
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	return tempDir, nil
}

// CopyEmbeddedFilesToDirectory copies files from an embedded filesystem to a specified directory.
//
// It walks through the embedded filesystem and writes each file to the destination directory,
// preserving the directory structure. Directories are created as needed.
func CopyEmbeddedFilesToDirectory(embeddedSchemas embed.FS, tempDir string) error {
	return fs.WalkDir(embeddedSchemas, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			content, readErr := embeddedSchemas.ReadFile(path)
			if readErr != nil {
				return readErr
			}

			destPath := filepath.Join(tempDir, path)

			// Create the directory if it doesn't exist
			if err = os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return err
			}

			return os.WriteFile(destPath, content, 0600)
		}
		return nil
	})
}
