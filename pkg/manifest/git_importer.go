package manifest

import (
	"os"
	"path/filepath"

	"github.com/google/go-jsonnet"
	"github.com/kartverket/skipctl/pkg/git"
)

type CachingFileImporter struct {
	cache map[string]jsonnet.Contents
}

func NewCachingFileImporter() *CachingFileImporter {
	return &CachingFileImporter{
		cache: make(map[string]jsonnet.Contents),
	}
}

func (fi *CachingFileImporter) Import(importedFrom, importedPath string) (jsonnet.Contents, string, error) {
	var absPath string
	if filepath.IsAbs(importedPath) {
		absPath = importedPath
	} else {
		absPath = filepath.Join(filepath.Dir(importedFrom), importedPath)
	}
	absPath = filepath.Clean(absPath)

	// Check if we already have this file cached
	if contents, exists := fi.cache[absPath]; exists {
		return contents, absPath, nil
	}

	// Read the file
	content, err := os.ReadFile(absPath)
	if err != nil {
		return jsonnet.Contents{}, "", err
	}

	// Create Contents instance and cache it
	contents := jsonnet.MakeContents(string(content))
	fi.cache[absPath] = contents

	return contents, absPath, nil
}

type GitImporter struct {
	Ref   string
	cache map[string]jsonnet.Contents
}

func NewGitImporter(ref string) *GitImporter {
	return &GitImporter{
		Ref:   ref,
		cache: make(map[string]jsonnet.Contents),
	}
}

func (gi *GitImporter) Import(importedFrom, importedPath string) (jsonnet.Contents, string, error) {
	// Resolve the path relative to importedFrom
	absPath := filepath.Join(filepath.Dir(importedFrom), importedPath)
	// Clean the path to resolve any .. or . components
	absPath = filepath.Clean(absPath)

	// Check if we already have this file cached
	if contents, exists := gi.cache[absPath]; exists {
		return contents, absPath, nil
	}

	// Fetch the content at the correct ref
	content, err := git.GetFileContentAtRef(absPath, gi.Ref)
	if err != nil {
		return jsonnet.Contents{}, "", err
	}

	// Create Contents instance and cache it
	contents := jsonnet.MakeContents(*content)
	gi.cache[absPath] = contents

	return contents, absPath, nil
}
