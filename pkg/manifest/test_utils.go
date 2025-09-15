package manifest

import (
	"path/filepath"
)

const filePermission = 0o600

func newTestDocument(content string, name string) *Document {
	return &Document{
		Name:        name,
		Content:     content,
		Permissions: filePermission,
		FromStdin:   true,
		Extension:   filepath.Ext(name),
	}
}
