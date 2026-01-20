package manifest

import (
	"path/filepath"
)

const FilePermission = 0o600

func NewTestDocument(content string, name string) *Document {
	return &Document{
		Name:        name,
		Content:     content,
		Permissions: FilePermission,
		FromStdin:   true,
		Extension:   filepath.Ext(name),
	}
}
