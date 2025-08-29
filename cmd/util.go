package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kartverket/skipctl/pkg/discovery"
	"github.com/spf13/cobra"
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
	var allFiles []string

	for _, suf := range suffixes {
		files, err := findFilesWithSuffix(directory, suf)
		if err != nil {
			return nil, err
		}
		allFiles = append(allFiles, files...)
	}

	return allFiles, nil
}

func findFilesWithSuffix(directory string, suffix string) ([]string, error) {
	var files []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == suffix {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
