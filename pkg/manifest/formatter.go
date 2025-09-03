package manifest

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/go-jsonnet/formatter"
	"github.com/kartverket/skipctl/pkg/constants"
)

func FormatJsonnet(filename string) error {
	rawContents, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	formatted, err := formatter.Format(filename, string(rawContents), formatter.DefaultOptions())
	if err != nil {
		return err
	}

	fileInfo, err := os.Stat(filename)
	if err != nil {
		return err
	}

	if err = os.WriteFile(filename, []byte(formatted), fileInfo.Mode().Perm()); err != nil {
		return err
	}

	return nil
}

func FormatManifest(filename string) error {
	switch filepath.Ext(filename) {
	case constants.ManifestSuffixJsonnet:
		return FormatJsonnet(filename)
	default:
		return fmt.Errorf("invalid file format in file %s", filename)
	}
}
