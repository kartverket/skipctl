package manifest

import (
	"bytes"
	"testing"

	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRendererWithLogger() (*Renderer, *bytes.Buffer) {
	var buf bytes.Buffer
	logger := logging.NewRawLoggerTo(&buf)
	renderer := NewRenderer(logger)

	return renderer, &buf
}

func TestRenderManifestValidJsonnet(t *testing.T) {
	validJsonnet := `
local port = 8080;
{
host: "localhost",
port: port,
ingress: [],
}
`
`
	expected := `{
   "host": "localhost",
   "ingress": [ ],
   "port": 8080
}
`

	doc := newTestDocument(validJsonnet, "valid.jsonnet")
	renderer, buf := newRendererWithLogger()
	err := renderer.RenderManifest(doc)
	got := buf.String()

	require.NoError(t, err, "expected no error for valid Jsonnet input")
	assert.Equal(t, expected, got, "rendered json did not match expected")
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
	inputYaml := `
application:
  host: localhost
  port: 8080
  spec:
    access-policies:
      enabled: true
`
	doc := newTestDocument(inputYaml, "input.yaml")
	renderer, buf := newRendererWithLogger()
	renderer.RenderManifest(doc)
	got := buf.String()

	assert.YAMLEq(t, inputYaml, got, "bruh")
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
