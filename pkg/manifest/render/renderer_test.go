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
	expected := `{
   "host": "localhost",
   "ingress": [ ],
   "port": 8080
}
`

	doc := newTestDocument(validJsonnet, "valid.jsonnet")
	renderer, buf := newRendererWithLogger()
	err := renderer.Render(doc)
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
	renderer, _ := newRendererWithLogger()
	res := renderer.Render(doc)
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
	err := renderer.Render(doc)
	got := buf.String()

	require.NoError(t, err, "expected no error for valid yaml input")

	assert.YAMLEq(t, inputYaml, got, "expected same yaml output as input")
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
	renderer, _ := newRendererWithLogger()
	res := renderer.Render(doc)
	require.Error(t, res, "expected error for invalid Yaml input")
}
func TestRenderManifestValidKustomize(t *testing.T) {
	// Use actual kustomize test directory
	kustomizePath := "../../testdata/kustomize/kustomization.yaml"

	doc := &Document{
		Name:        kustomizePath,
		Extension:   "kustomization.yaml",
		Permissions: filePermission,
		FromStdin:   false,
	}

	renderer, buf := newRendererWithLogger()
	err := renderer.Render(doc)

	require.NoError(t, err, "expected no error for valid Kustomize")
	got := buf.String()
	assert.NotEmpty(t, got, "expected kustomize output")
	assert.Contains(t, got, "kind: Application", "output should contain Application")
}
func TestRenderManifestInvalidKustomize(t *testing.T) {
	// Point to a directory that doesn't exist
	doc := &Document{
		Name:        "/nonexistent/path/kustomization.yaml",
		Extension:   "kustomization.yaml",
		Permissions: filePermission,
		FromStdin:   false,
	}

	renderer, _ := newRendererWithLogger()
	err := renderer.Render(doc)

	require.Error(t, err, "expected error for non-existent kustomization directory")
	assert.Contains(t, err.Error(), "kustomize build", "error should mention kustomize build")
}
