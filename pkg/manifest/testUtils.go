package manifest

import (
	"os"
	"path/filepath"
)

const filePermission = 0o600

func writeContentToTmpDir(dir string, content string, filename string) (string, error) {
	path := dir + string(os.PathSeparator) + filename

	err := os.WriteFile(path, []byte(content), filePermission)
	if err != nil {
		return "", err
	}
	return path, err
}

func newTestDocument(content string, name string) *Document {
	return &Document{
		Name:        name,
		Content:     content,
		Permissions: filePermission,
		FromStdin:   true,
		Extension:   filepath.Ext(name),
	}
}
