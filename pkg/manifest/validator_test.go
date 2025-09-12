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

	doc := newTestDocument(validJsonnetManifest, "valid.jsonnet")
	validator := NewValidator(t.TempDir())
	err := validator.ValidateManifest(doc)
	result := validator.GetResults()

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

	doc := newTestDocument(validJsonnet, "valid.jsonnet")
	validator := NewValidator(t.TempDir())
	err := validator.ValidateManifest(doc)
	result := validator.GetResults()

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

	doc := newTestDocument(invalidJsonnet, "invalid.jsonnet")
	validator := NewValidator(t.TempDir())
	err := validator.ValidateManifest(doc)
	result := validator.GetResults()

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

	doc := newTestDocument(validYamlManifest, "valid.yaml")
	validator := NewValidator(t.TempDir())
	err := validator.ValidateManifest(doc)
	result := validator.GetResults()

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

	doc := newTestDocument(validYaml, "valid.yaml")
	validator := NewValidator(t.TempDir())
	err := validator.ValidateManifest(doc)
	result := validator.GetResults()

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
kind Application
metadata
  			name: valid-manifest
  namespace: devex
spec:
  image: "kartverket/example"
`

	doc := newTestDocument(invalidYaml, "invalid.yaml")
	validator := NewValidator(t.TempDir())
	_ = validator.ValidateManifest(doc) // TODO does not return error, why?
	result := validator.GetResults()
	hasError := result.HasValidationFailed()

	// require.Error(t, err, "expected error for invalid Yaml input")
	assert.True(t, hasError, "expected error for inavlid yaml")
	assert.Equal(t, 1, result.ErrorCount, "unexpected result.ErrorCount")
	assert.Equal(t, 0, result.InvalidCount, "unexpected result.InvalidCount")
	assert.Equal(t, 0, result.SkippedCount, "unexpected result.SkippedCount")
	assert.Equal(t, 0, result.ValidCount, "unexpected result.ValidCount")
}
