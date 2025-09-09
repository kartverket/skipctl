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

	tmp := t.TempDir()
	filename := writeContentToTmpDir(tmp, validJsonnet, "valid.jsonnet")

	res := FormatManifest(filename)
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

	tmp := t.TempDir()
	filename := writeContentToTmpDir(tmp, invalidJsonnet, "invalid.jsonnet")

	res := FormatManifest(filename)
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

	tmp := t.TempDir()
	filename := writeContentToTmpDir(tmp, validYaml, "valid.yaml")

	res := FormatManifest(filename)
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

	tmp := t.TempDir()
	filename := writeContentToTmpDir(tmp, invalidYaml, "invalid.yaml")

	res := FormatManifest(filename)
	require.Error(t, res, "expected error for invalid Yaml input")
}
