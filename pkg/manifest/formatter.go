package manifest

import (
	"log"
	"os"

	"github.com/google/go-jsonnet/formatter"
)

func FormatManifest(filepath string) (string, error) {

	byteOut, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalf("Error reading file %v", err)
	}
	outStr := string(byteOut)
	jsonOut, err := formatter.Format(filepath, outStr, formatter.DefaultOptions())
	if err != nil {
		log.Fatalf("Error formatting .jsonnet file %v", err)
	}
	err = os.WriteFile(filepath, []byte(jsonOut), 0677)
	if err != nil {
		log.Fatalf("Error writing to file %v", err)
	}

	return jsonOut, err

}
