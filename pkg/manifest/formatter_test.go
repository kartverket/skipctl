package manifest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatManifestValidJsonnet(t *testing.T) {
	validJsonnet := `
{
host: "localhost",
port: 8080,
ingress: [],
}
`

	doc := newTestDocument(validJsonnet, "valid.jsonnet")
	res := FormatManifest(doc)
	require.NoError(t, res, "expected no error for valid Jsonnet input")
}

func TestFormatManifestInvalidJsonnet(t *testing.T) {
	invalidJsonnet := `
{
host: localhost
port: 8080,
ingress = []
}
`

	doc := newTestDocument(invalidJsonnet, "invalid.jsonnet")
	res := FormatManifest(doc)
	require.Error(t, res, "expected error for invalid Jsonnet input")
}

func TestFormatManifestValidYaml(t *testing.T) {
	validYaml := `
application:
  host: localhost
  port: 8080
  ingress:
    - item
`

	doc := newTestDocument(validYaml, "valid.yaml")
	res := FormatManifest(doc)
	require.NoError(t, res, "expected no error for valid Yaml input")
}

func TestFormatManifestInvalidYaml(t *testing.T) {
	invalidYaml := `
application:
		host: localhost
port: 8080
	ingress
}
`

	doc := newTestDocument(invalidYaml, "invalid.yaml")
	res := FormatManifest(doc)
	require.Error(t, res, "expected error for invalid Yaml input")
}
