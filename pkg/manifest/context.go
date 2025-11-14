package manifest

// Context holds shared resources for manifest operations.
// It lazily loads the git filesystem only when needed.
type Context struct {
	ref string
}

// NewContext creates a new context for manifest operations.
func NewContext(ref string) *Context {
	return &Context{
		ref: ref,
	}
}

// Ref returns the git reference.
func (c *Context) Ref() string {
	return c.ref
}
