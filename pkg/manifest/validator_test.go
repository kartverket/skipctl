package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateManifestJsonnetValid(t *testing.T) {
	// The syntax and content is valid.
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

	require.NoError(t, err, "ValidateManifest should not return an error for valid Jsonnet input")
	assert.Equal(t, 0, result.ErrorCount, "unexpected result.ErrorCount")
	assert.Equal(t, 0, result.InvalidCount, "unexpected result.InvalidCount")
	assert.Equal(t, 0, result.SkippedCount, "unexpected result.SkippedCount")
	assert.Equal(t, 1, result.ValidCount, "unexpected result.ValidCount")
}

func TestValidateManifestJsonnetInvalid(t *testing.T) {
	// The syntax is valid but the spec is missing port. This manifest is invalid.
	validJsonnet := `
{
  apiVersion: "skiperator.kartverket.no/v1alpha1",
  kind: "Application",
  metadata: {
    name: "valid-manifest",
    namespace: "devex",
  },
  spec: {
    image: "kartverket/example",
  },
}
`

	tmp := t.TempDir()
	filename := writeContentToTmpDir(tmp, validJsonnet, "valid.jsonnet")
	result, err := NewValidator(tmp).ValidateManifest(filename)

	require.NoError(t, err, "ValidateManifest should not return an error for valid Jsonnet input")
	assert.Equal(t, 0, result.ErrorCount, "unexpected result.ErrorCount")
	assert.Equal(t, 1, result.InvalidCount, "unexpected result.InvalidCount")
	assert.Equal(t, 0, result.SkippedCount, "unexpected result.SkippedCount")
	assert.Equal(t, 0, result.ValidCount, "unexpected result.ValidCount")
}

func TestValidateManifestJsonnetSyntaxError(t *testing.T) {
	// This document contains syntax errors and should fail validation
	invalidJsonnet := `
	{
  apiVersion: "skiperator.kartverket.no/v1alpha1",
  kind: "Application",
  metadata:
    name: "valid-manifest",
    namespace: "devex"
  },
  spec: {
    image: "kartverket/example",
    port: 8080,
  },
}
`

	tmp := t.TempDir()
	filename := writeContentToTmpDir(tmp, invalidJsonnet, "invalid.jsonnet")
	result, err := NewValidator(tmp).ValidateManifest(filename)

	require.Error(t, err, "expected error for invalid Jsonnet input")

	assert.Equal(t, 1, result.ErrorCount, "unexpected result.ErrorCount")
	assert.Equal(t, 0, result.InvalidCount, "unexpected result.InvalidCount")
	assert.Equal(t, 0, result.SkippedCount, "unexpected result.SkippedCount")
	assert.Equal(t, 0, result.ValidCount, "unexpected result.ValidCount")
}

func TestValidateManifestYamlValid(t *testing.T) {
	// this is a valid manifest
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

	require.NoError(t, err, "ValidateManifest should not return an error for valid yaml input")
	assert.Equal(t, 0, result.ErrorCount, "unexpected result.ErrorCount")
	assert.Equal(t, 0, result.InvalidCount, "unexpected result.InvalidCount")
	assert.Equal(t, 0, result.SkippedCount, "unexpected result.SkippedCount")
	assert.Equal(t, 1, result.ValidCount, "unexpected result.ValidCount")
}

func TestValidateManifestYamlInvalid(t *testing.T) {
	// this is valid yaml syntax but spec is missing port field, should fail validation
	validYaml := `
apiVersion: skiperator.kartverket.no/v1alpha1
kind: Application
metadata:
  name: valid-manifest
  namespace: devex
spec:
  image: "kartverket/example"
`

	tmp := t.TempDir()
	filename := writeContentToTmpDir(tmp, validYaml, "valid.yaml")
	result, err := NewValidator(tmp).ValidateManifest(filename)

	require.NoError(t, err, "expected no error for valid Yaml input")

	assert.Equal(t, 0, result.ErrorCount, "unexpected result.ErrorCount")
	assert.Equal(t, 1, result.InvalidCount, "unexpected result.InvalidCount")
	assert.Equal(t, 0, result.SkippedCount, "unexpected result.SkippedCount")
	assert.Equal(t, 0, result.ValidCount, "unexpected result.ValidCount")
}

func TestValidateManifestYamlSyntaxError(t *testing.T) {
	// This document has syntax errors, should fail validation
	invalidYaml := `
	apiVersion: skiperator.kartverket.no/v1alpha1
kind: Application
metadata
  			name: valid-manifest
  namespace: devex
spec:
  image: "kartverket/example"
`

	tmp := t.TempDir()
	filename := writeContentToTmpDir(tmp, invalidYaml, "invalid.yaml")
	result, err := NewValidator(tmp).ValidateManifest(filename)

	require.Error(t, err, "expected error for invalid Yaml input")

	assert.Equal(t, 1, result.ErrorCount, "unexpected result.ErrorCount")
	assert.Equal(t, 0, result.InvalidCount, "unexpected result.InvalidCount")
	assert.Equal(t, 0, result.SkippedCount, "unexpected result.SkippedCount")
	assert.Equal(t, 0, result.ValidCount, "unexpected result.ValidCount")
}
