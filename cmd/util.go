package cmd

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kartverket/skipctl/pkg/discovery"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"
)

var activeAPIServer discovery.APIServer

func ValidateAPIServerName(_ *cobra.Command, _ []string) {
	if len(apiServer) == 0 {
		log.Error("no api server specified, exiting")
		os.Exit(1)
	}

	var matchFound = false
	for _, server := range apiServers {
		if strings.EqualFold(apiServer, server.Name) {
			matchFound = true
			activeAPIServer = server
			break
		}
	}

	if !matchFound {
		var names []string
		for _, server := range apiServers {
			names = append(names, strings.ToLower(server.Name))
		}

		log.Error("unknown api server - please pick another supported", "specified", apiServer, "supported", names)
		os.Exit(1)
	}
}

func findFilesWithSuffixes(directory string, suffixes []string) ([]string, error) {
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
