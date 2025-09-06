package crd

import "embed"

// Schemas contains the embedded JSON schema files for CRD validation.
//
//go:embed schemas/*.json
var Schemas embed.FS
