package manifest

import (
	"os"
	"path/filepath"

	"github.com/google/go-jsonnet"
	"github.com/kartverket/skipctl/pkg/git"
)

// Shared cache for all importers
type ImportCache struct {
	cache map[string]jsonnet.Contents
}

func NewImportCache() *ImportCache {
	return &ImportCache{
		cache: make(map[string]jsonnet.Contents),
	}
}

type FileImporter struct {
	sharedCache *ImportCache
}

func NewFileImporter(cache *ImportCache) *FileImporter {
	if cache == nil {
		cache = NewImportCache()
	}
	return &FileImporter{
		sharedCache: cache,
	}
}

func (fi *FileImporter) Import(importedFrom, importedPath string) (jsonnet.Contents, string, error) {
	var absPath string
	if filepath.IsAbs(importedPath) {
		absPath = importedPath
	} else {
		absPath = filepath.Join(filepath.Dir(importedFrom), importedPath)
	}
	absPath = filepath.Clean(absPath)

	// Check if we already have this file cached
	if contents, exists := fi.sharedCache.cache[absPath]; exists {
		return contents, absPath, nil
	}

	// Read the file
	content, err := os.ReadFile(absPath)
	if err != nil {
		return jsonnet.Contents{}, "", err
	}

	// Create Contents instance and cache it
	contents := jsonnet.MakeContents(string(content))
	fi.sharedCache.cache[absPath] = contents

	return contents, absPath, nil
}

type GitFileImporter struct {
	Ref         string
	sharedCache *ImportCache
}

func NewGitFileImporter(ref string, cache *ImportCache) *GitFileImporter {
	if cache == nil {
		cache = NewImportCache()
	}
	return &GitFileImporter{
		Ref:         ref,
		sharedCache: cache,
	}
}

func (gi *GitFileImporter) Import(importedFrom, importedPath string) (jsonnet.Contents, string, error) {
	// Resolve the path relative to importedFrom
	absPath := filepath.Join(filepath.Dir(importedFrom), importedPath)
	absPath = filepath.Clean(absPath)

	// Check if we already have this file cached
	if contents, exists := gi.sharedCache.cache[gi.Ref+":"+absPath]; exists {
		return contents, absPath, nil
	}

	// Fetch the content at the correct ref
	content, err := git.GetFileContentAtRef(absPath, gi.Ref)
	if err != nil {
		return jsonnet.Contents{}, "", err
	}

	// Create Contents instance and cache it
	contents := jsonnet.MakeContents(*content)
	gi.sharedCache.cache[gi.Ref+":"+absPath] = contents

	return contents, absPath, nil
}
