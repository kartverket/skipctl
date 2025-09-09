package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateManifestValidJsonnetManifest(t *testing.T) {
	validJsonnetManifest := `
{
  apiVersion: "skiperator.kartverket.no/v1alpha1",
  kind: "Application",
  metadata: {
    name: "valid-manifest",
    namespace: "devex",
  },
  spec: {
    image: "kartverket/example",
    port: 8080,
    replicas: {
      min: 1,
      max: 5,
    },
  },
}
`

	tmp := t.TempDir()
	filename := writeContentToTmpDir(tmp, validJsonnetManifest, "valid.jsonnet")

	result, err := NewValidator(tmp).ValidateManifest(filename)

	assert.NoError(t, err, "ValidateManifest should not return an error for valid Jsonnet input")
	assert.Equal(t, 0, result.ErrorCount)
	assert.Equal(t, 0, result.InvalidCount)
	assert.Equal(t, 0, result.SkippedCount)
	assert.Equal(t, 1, result.ValidCount)
}

func TestValidateManifestValidJsonnet(t *testing.T) {
	validJsonnet := `
{
host: "localhost",
port: 8080,
ingress: [],
}
`

	tmp := t.TempDir()
	filename := writeContentToTmpDir(tmp, validJsonnet, "valid.jsonnet")

	result, err := NewValidator(tmp).ValidateManifest(filename)

	assert.NoError(t, err, "ValidateManifest should not return an error for valid Jsonnet input")
	assert.Equal(t, 0, result.ErrorCount)
	assert.Equal(t, 1, result.InvalidCount)
	assert.Equal(t, 0, result.SkippedCount)
	assert.Equal(t, 0, result.ValidCount)
}

func TestValidateManifestInvalidJsonnet(t *testing.T) {
	invalidJsonnet := `
{
host: localhost
port: 8080,
ingress = []
}
`

	tmp := t.TempDir()
	filename := writeContentToTmpDir(tmp, invalidJsonnet, "invalid.jsonnet")

	result, err := NewValidator(tmp).ValidateManifest(filename)

	assert.Error(t, err, "expected error for invalid Jsonnet input")

	assert.Equal(t, 1, result.ErrorCount)
	assert.Equal(t, 0, result.InvalidCount)
	assert.Equal(t, 0, result.SkippedCount)
	assert.Equal(t, 0, result.ValidCount)
}

func TestValidateManifestValidYamlManifest(t *testing.T) {
	validYamlManifest := `
apiVersion: skiperator.kartverket.no/v1alpha1
kind: Application
metadata:
  name: valid-manifest
  namespace: devex
spec:
  image: "kartverket/example"
  port: 8080
  replicas:
    min: 1
    max: 5
`

	tmp := t.TempDir()
	filename := writeContentToTmpDir(tmp, validYamlManifest, "valid.yaml")

	result, err := NewValidator(tmp).ValidateManifest(filename)

	assert.NoError(t, err, "ValidateManifest should not return an error for valid yaml input")
	assert.Equal(t, 0, result.ErrorCount)
	assert.Equal(t, 0, result.InvalidCount)
	assert.Equal(t, 0, result.SkippedCount)
	assert.Equal(t, 1, result.ValidCount)
}

func TestValidateManifestValidYaml(t *testing.T) {
	validYaml := `
application:
  host: localhost
  port: 8080
  ingress:
    - item
`

	tmp := t.TempDir()
	filename := writeContentToTmpDir(tmp, validYaml, "valid.yaml")

	result, err := NewValidator(tmp).ValidateManifest(filename)

	assert.NoError(t, err, "expected no error for valid Yaml input")

	assert.Equal(t, 0, result.ErrorCount)
	assert.Equal(t, 1, result.InvalidCount)
	assert.Equal(t, 0, result.SkippedCount)
	assert.Equal(t, 0, result.ValidCount)
}

func TestValidateManifestInvalidYaml(t *testing.T) {
	invalidYaml := `
application:
		host: localhost
port: 8080
	ingress
}
`

	tmp := t.TempDir()
	filename := writeContentToTmpDir(tmp, invalidYaml, "invalid.yaml")

	result, err := NewValidator(tmp).ValidateManifest(filename)

	assert.Error(t, err, "expected error for invalid Yaml input")

	assert.Equal(t, 1, result.ErrorCount)
	assert.Equal(t, 0, result.InvalidCount)
	assert.Equal(t, 0, result.SkippedCount)
	assert.Equal(t, 0, result.ValidCount)
}
