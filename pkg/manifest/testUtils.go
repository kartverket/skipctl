package manifest

import (
	"os"
	"path/filepath"
)

func writeContentToTmpDir(dir string, content string, filename string) string {
	path := dir + string(os.PathSeparator) + filename

	os.WriteFile(path, []byte(content), 0644)

	return path
}

func newTestDocument(content string, name string) *Document {
	return &Document{
		Name:        name,
		Content:     content,
		Permissions: 0644,
		FromStdin:   true,
		Extension:   filepath.Ext(name),
	}
}
