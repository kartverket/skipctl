package manifest

import (
	"fmt"
	"sync"

	"github.com/kartverket/skipctl/pkg/git"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// Context holds shared resources for manifest operations.
// It lazily loads the git filesystem only when needed.
type Context struct {
	ref       string
	gitFS     filesys.FileSystem
	currentFS filesys.FileSystem
	mu        sync.Mutex
	gitFSErr  error
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

// GitFS returns the in-memory filesystem at the git ref.
// It loads the filesystem lazily on first access.
func (c *Context) GitFS() (filesys.FileSystem, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.gitFS != nil {
		return c.gitFS, nil
	}

	if c.gitFSErr != nil {
		return nil, c.gitFSErr
	}

	fs, err := git.GetInMemoryFilesystemAt(c.ref)
	if err != nil {
		c.gitFSErr = fmt.Errorf("load git filesystem at %s: %w", c.ref, err)
		return nil, c.gitFSErr
	}

	c.gitFS = fs
	return c.gitFS, nil
}

// Ref returns the git reference.
func (c *Context) Ref() string {
	return c.ref
}
