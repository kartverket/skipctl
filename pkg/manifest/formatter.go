package manifest

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/google/go-jsonnet/formatter"
	"github.com/kartverket/skipctl/pkg/constants"
)

func FormatJsonnet(filename string) error {
	byteOut, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	outStr := string(byteOut)
	jsonOut, err := formatter.Format(filename, outStr, formatter.DefaultOptions())
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, []byte(jsonOut), 0600)
	if err != nil {
		return err
	}

	return err
}

func FormatManifest(filename string) error {
	switch filepath.Ext(filename) {
	case constants.ManifestSuffixJsonnet:
		return FormatJsonnet(filename)
	default:
		return errors.New("invalid file format")
	}
}
