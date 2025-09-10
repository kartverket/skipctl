package manifest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderManifestValidJsonnet(t *testing.T) {
	validJsonnet := `
{
host: "localhost",
port: 8080,
ingress: [],
}
`

	doc := newTestDocument(validJsonnet, "valid.jsonnet")
	res := NewRenderer().RenderManifest(doc)
	require.NoError(t, res, "expected no error for valid Jsonnet input")
}

func TestRenderManifestInvalidJsonnet(t *testing.T) {
	invalidJsonnet := `
{
host: localhost
port: 8080,
ingress = []
}
`

	doc := newTestDocument(invalidJsonnet, "invalid.jsonnet")
	res := NewRenderer().RenderManifest(doc)
	require.Error(t, res, "expected error for invalid Jsonnet input")
}

func TestRenderManifestValidYaml(t *testing.T) {
	validYaml := `
application:
  host: localhost
  port: 8080
  ingress:
    - item
`

	doc := newTestDocument(validYaml, "valid.yaml")
	res := NewRenderer().RenderManifest(doc)
	require.NoError(t, res, "expected no error for valid Yaml input")
}

func TestRenderManifestInvalidYaml(t *testing.T) {
	invalidYaml := `
application:
		host: localhost
port: 8080
	ingress
}
`

	doc := newTestDocument(invalidYaml, "invalid.yaml")
	res := NewRenderer().RenderManifest(doc)
	require.Error(t, res, "expected error for invalid Yaml input")
}
