package crd

import "embed"

// Schemas contains the embedded JSON schema files for CRD validation.
//
//go:embed schemas/*.json
var Schemas embed.FS

// ListSchemas lists the names of all embedded schema files.
func ListSchemas() ([]string, error) {
	entries, err := Schemas.ReadDir("schemas")
	if err != nil {
		return nil, err
	}

	var schemaFiles []string
	for _, entry := range entries {
		if !entry.IsDir() {
			schemaFiles = append(schemaFiles, entry.Name())
		}
	}
	return schemaFiles, nil
}
