package manifest

import (
	"fmt"

	"github.com/kartverket/skipctl/pkg/utils"
)

// Source provides access to previous versions of manifests.
type Source interface {
	GetPreviousDocument(current *Document) (*Document, error)
	Reference() string
}

// GitSource retrieves previous versions from git.
type GitSource struct {
	ref string
}

// NewGitSource creates a new GitSource.
func NewGitSource(ref string) *GitSource {
	return &GitSource{ref: ref}
}

// GetPreviousDocument retrieves a document from git at the specified ref.
func (s *GitSource) GetPreviousDocument(current *Document) (*Document, error) {
	return current.AtRef(s.ref)
}

// Reference returns the git reference.
func (s *GitSource) Reference() string {
	return s.ref
}

// DirectorySource retrieves previous versions from a filesystem directory.
type DirectorySource struct {
	baseDir   string
	originDir string
}

// NewDirectorySource creates a new DirectorySource.
func NewDirectorySource(baseDir string, originDir string) *DirectorySource {
	return &DirectorySource{baseDir: baseDir, originDir: originDir}
}

// GetPreviousDocument retrieves a document from the base directory.
func (s *DirectorySource) GetPreviousDocument(current *Document) (*Document, error) {
	rebased := utils.RebasePath(current.Name, s.originDir, s.baseDir)

	documents, err := FromFiles([]string{rebased})
	if err != nil {
		return nil, err
	}

	if len(documents) == 0 {
		return nil, fmt.Errorf("did not find any documents matching %s", rebased)
	}

	doc := documents[0]

	return doc, nil
}

// Reference returns the base directory path.
func (s *DirectorySource) Reference() string {
	return s.baseDir
}
