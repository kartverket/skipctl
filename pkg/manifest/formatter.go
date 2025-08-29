package manifest

import (
	"log"
	"os"
	"path/filepath"

	"github.com/google/go-jsonnet/formatter"
	"github.com/kartverket/skipctl/pkg/constants"
)

func FormatJsonnet(filename string) (string, error) {
	byteOut, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading file %v", err)
	}
	outStr := string(byteOut)
	jsonOut, err := formatter.Format(filename, outStr, formatter.DefaultOptions())
	if err != nil {
		log.Fatalf("Error formatting .jsonnet file %v", err)
	}
	err = os.WriteFile(filename, []byte(jsonOut), 0600)
	if err != nil {
		log.Fatalf("Error writing to file %v", err)
	}

	return jsonOut, err
}

func FormatManifest(filename string) (string, error) {
	switch filepath.Ext(filename) {
	case constants.Suffixes[0]:
		return FormatJsonnet(filename)
	default:
		return "", nil
	}
}
