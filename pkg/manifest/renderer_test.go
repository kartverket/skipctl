package manifest

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeContentToTmpDir(dir string, content string, filename string) string {
	path := dir + string(os.PathSeparator) + filename

	os.WriteFile(path, []byte(content), 0644)

	return path
}

func TestRenderManifestValidJsonnet(t *testing.T) {
	validJsonnet := `
{
host: "localhost",
port: 8080,
ingress: [],
}
`

	tmp := t.TempDir()
	filename := writeContentToTmpDir(tmp, validJsonnet, "valid.jsonnet")

	res := NewRenderer().RenderManifest(filename)
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

	tmp := t.TempDir()
	filename := writeContentToTmpDir(tmp, invalidJsonnet, "invalid.jsonnet")

	res := NewRenderer().RenderManifest(filename)
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

	tmp := t.TempDir()
	filename := writeContentToTmpDir(tmp, validYaml, "valid.yaml")

	res := NewRenderer().RenderManifest(filename)
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

	tmp := t.TempDir()
	filename := writeContentToTmpDir(tmp, invalidYaml, "invalid.yaml")

	res := NewRenderer().RenderManifest(filename)
	require.Error(t, res, "expected error for invalid Yaml input")
}
