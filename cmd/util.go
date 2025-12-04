package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/kartverket/skipctl/pkg/discovery"
	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/kartverket/skipctl/pkg/utils"
	"github.com/spf13/cobra"
)

var activeAPIServer discovery.APIServer

func ValidateAPIServerName(_ *cobra.Command, _ []string) error {
	if len(apiServer) == 0 {
		return errors.New("no api server specified")
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

		return fmt.Errorf("unknown api server '%s'  - please pick another supported: %s", apiServer, names)
	}

	return nil
}

var (
	// HEAD with allowed suffixes: ~N, ^, ^N, optionally @{N}
	reHead = regexp.MustCompile(`^HEAD(?:~[0-9]+|\^[0-9]*|@\{\d+\})?$`)

	// 7–40 hex SHA
	reSHA = regexp.MustCompile(`^[0-9a-fA-F]{7,40}$`)

	// Basic branch/tag-like names (broad)
	reBranch = regexp.MustCompile(`^[A-Za-z0-9._/-]+$`)
)

func isAllHex(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			continue
		}
		return false
	}
	return true
}

func IsValidCommitRef(ref string) bool {
	// 1) HEAD and suffixes
	if reHead.MatchString(ref) {
		return true
	}

	// 2) 7–40 hex
	if reSHA.MatchString(ref) {
		return true
	}

	// 3) Branch-like names, but exclude:
	//    - exact "HEAD" handled above (already excluded)
	//    - lowercase "head" (explicitly reject)
	//    - any pure-hex string of any length (so 6, 41, 50, etc. don't sneak in)
	if reBranch.MatchString(ref) {
		if ref == "head" {
			return false
		}
		if isAllHex(ref) {
			return false
		}
		return true
	}
	return false
}

func determineDocuments(pathFlag string, args []string, suffixes []string) ([]*manifest.Document, error) {
	var manifestFiles []*manifest.Document
	var err error

	// priority: -p flag > positional arg > current directory
	filePath := "."
	if pathFlag != "" {
		filePath = pathFlag
	} else if len(args) > 0 {
		filePath = args[0]
	}

	if filePath == "-" {
		manifestFiles, err = manifest.FromStdin()
	} else {
		var filenames []string
		filenames, err = utils.FindFilesWithSuffixes(filePath, suffixes)
		if err == nil {
			manifestFiles, err = manifest.FromFiles(filenames)
		}
	}

	return manifestFiles, err
}
