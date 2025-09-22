package manifest

import (
	"path/filepath"

	"github.com/google/go-jsonnet"
	"github.com/kartverket/skipctl/pkg/utils"
)

type GitImporter struct {
	Ref string
}

func NewGitImporter(ref string) *GitImporter {
	return &GitImporter{Ref: ref}
}

func (gi *GitImporter) Import(importedFrom, importedPath string) (jsonnet.Contents, string, error) {
	// Resolve the path relative to importedFrom
	absPath := filepath.Join(filepath.Dir(importedFrom), importedPath)
	// Fetch the content at the correct ref
	content, err := utils.GetFileContentFromRef(absPath, gi.Ref)
	if err != nil {
		return jsonnet.Contents{}, "", err
	}
	return jsonnet.MakeContents(*content), absPath, nil
}
