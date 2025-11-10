package manifest

import (
	"fmt"
	"os"
	"sync"

	"github.com/kartverket/skipctl/pkg/git"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// Context holds shared resources for manifest operations.
// It lazily loads the git filesystem only when needed.
type Context struct {
	ref       string
	gitDir    string
	currentFS filesys.FileSystem
	mu        sync.Mutex
	gitDirErr error
}

// NewContext creates a new context for manifest operations.
func NewContext(ref string) *Context {
	return &Context{
		ref:       ref,
		currentFS: filesys.MakeFsOnDisk(),
	}
}

// CurrentFS returns the filesystem for the current working directory.
func (c *Context) CurrentFS() filesys.FileSystem {
	return c.currentFS
}

// GitDir returns the temporary directory with git ref contents.
// It loads the directory lazily on first access.
func (c *Context) GitDir() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.gitDir != "" {
		return c.gitDir, nil
	}

	if c.gitDirErr != nil {
		return "", c.gitDirErr
	}

	dir, err := git.GetFilesystemAt(c.ref)
	if err != nil {
		c.gitDirErr = fmt.Errorf("load git filesystem at %s: %w", c.ref, err)
		return "", c.gitDirErr
	}

	c.gitDir = dir
	return c.gitDir, nil
}

// CleanUp removes the temporary git directory if it was created.
func (c *Context) CleanUp() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.gitDir != "" {
		os.RemoveAll(c.gitDir)
		c.gitDir = ""
	}
}

// Ref returns the git reference.
func (c *Context) Ref() string {
	return c.ref
}
