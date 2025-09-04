package utils

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"go.yaml.in/yaml/v4"
)

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
