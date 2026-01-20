package format_test

import (
	"testing"

	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/kartverket/skipctl/pkg/manifest/format"
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

	doc := manifest.NewTestDocument(validJsonnet, "valid.jsonnet")
	res := format.Manifest(doc)
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

	doc := manifest.NewTestDocument(invalidJsonnet, "invalid.jsonnet")
	res := format.Manifest(doc)
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

	doc := manifest.NewTestDocument(validYaml, "valid.yaml")
	res := format.Manifest(doc)
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

	doc := manifest.NewTestDocument(invalidYaml, "invalid.yaml")
	res := format.Manifest(doc)
	require.Error(t, res, "expected error for invalid Yaml input")
}

func TestFormatManifestValidLibsonnet(t *testing.T) {
	validLibsonnet := `
{
host: "localhost",
port: 8080,
ingress: [],
}
`

	doc := manifest.NewTestDocument(validLibsonnet, "valid.libsonnet")
	res := format.Manifest(doc)
	require.NoError(t, res, "expected no error for valid Libsonnet input")
}

func TestFormatManifestInvalidLibsonnet(t *testing.T) {
	invalidLibsonnet := `
{
host: localhost
port: 8080,
ingress = []
}
`

	doc := manifest.NewTestDocument(invalidLibsonnet, "invalid.libsonnet")
	res := format.Manifest(doc)
	require.Error(t, res, "expected error for invalid Libsonnet input")
}
