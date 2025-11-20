package manifest

import (
	"testing"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/diff"
	"github.com/stretchr/testify/require"
)

// Simple function to count diff types in a manifest diff
func countDiffTypes(diffs []*diff.ManifestDiff) (int, int, int) {
	var insertions, deletions, equals int
	for _, d := range diffs {
		switch d.Type {
		case constants.Insertion:
			insertions++
		case constants.Deletion:
			deletions++
		case constants.Equals:
			equals++
		}
	}
	return insertions, deletions, equals
}

func TestDiffManifestJsonnetEqual(t *testing.T) {
	jsonnetManifest1 := `
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
	jsonnetManifest2 := `
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
	// Mock diffs with two equal manifests
	_, hasChanges := diff.MainDiff(jsonnetManifest1, jsonnetManifest2)
	require.False(t, hasChanges, "unepected should have zero changes but registered some")
}

func TestDiffManifestYamlEqual(t *testing.T) {
	yamlManifest1 := `
apiVersion: skiperator.kartverket.no/v1alpha1
kind: Application
metadata:
  name: valid-manifest
  namespace: devex
spec:
  image: kartverket/example
  port: 8080
  replicas:
    min: 1
    max: 5
`
	yamlManifest2 := `
apiVersion: skiperator.kartverket.no/v1alpha1
kind: Application
metadata:
  name: valid-manifest
  namespace: devex
spec:
  image: kartverket/example
  port: 8080
  replicas:
    min: 1
    max: 5
`
	// Mock diffs with two equal manifests
	_, hasChanges := diff.MainDiff(yamlManifest1, yamlManifest2)
	require.False(t, hasChanges, "unepected should have zero changes but registered some")
}

func TestDiffManifestJsonnetUnequal(t *testing.T) {
	jsonnetManifest1 := `{
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
}`
	jsonnetManifest2 := `{
  apiVersion: "skiperator.kartverket.no/v1alpha1",
  kind: "Application",
  metadata: {
    name: "valid-manifest",
    namespace: "devex",
  },
  spec: {
    image: "kartverket/example:v2",
    port: 8080,
    replicas: {
      min: 2,
      max: 10,
    },
  },
}`
	// Mock diffs with two different manifests
	diffs, hasChanges := diff.MainDiff(jsonnetManifest1, jsonnetManifest2)
	require.True(t, hasChanges, "expected changes but none were registered")
	// Count diff types, this should be 3 insertions and deletions from these manifests
	insertionCount, deletionCount, equalCount := countDiffTypes(diffs)
	expectedCountInsDel := 3
	expectedCountEq := 16 - expectedCountInsDel
	require.Equal(t, expectedCountInsDel, insertionCount)
	require.Equal(t, expectedCountInsDel, deletionCount)
	require.Equal(t, expectedCountEq, equalCount)
}

func TestDiffManifestYamlUnequal(t *testing.T) {
	yamlManifest1 := `apiVersion: skiperator.kartverket.no/v1alpha1
kind: Application
metadata:
  name: valid-manifest
  namespace: devex
spec:
  image: kartverket/example
  port: 8080
  replicas:
    min: 3
    max: 15`
	yamlManifest2 := `apiVersion: skiperator.kartverket.no/v1alpha1
kind: Application
metadata:
  name: valid-manifest
  namespace: production
spec:
  image: kartverket/example1
  port: 8080
  replicas:
    min: 1
    max: 15`

	// Mock diffs with two different manifests
	diffs, hasChanges := diff.MainDiff(yamlManifest1, yamlManifest2)
	require.True(t, hasChanges, "expected changes but none were registered")
	// Count diff types, this should be 3 insertions and deletions from these manifests
	insertionCount, deletionCount, equalCount := countDiffTypes(diffs)
	expectedCountInsDel := 3
	expectedCountEq := 11 - expectedCountInsDel
	require.Equal(t, expectedCountInsDel, insertionCount)
	require.Equal(t, expectedCountInsDel, deletionCount)
	require.Equal(t, expectedCountEq, equalCount)
}

func TestDiffManifestYamlNewManifest(t *testing.T) {
	yamlManifest1 := ``
	yamlManifest2 := `apiVersion: skiperator.kartverket.no/v1alpha1
kind: Application
metadata:
  name: valid-manifest
  namespace: production
spec:
  image: kartverket/example1
  port: 8080
  replicas:
    min: 1
    max: 15`

	// Mock diffs with two different manifests
	diffs, hasChanges := diff.MainDiff(yamlManifest1, yamlManifest2)
	require.True(t, hasChanges, "expected changes but none were registered")
	// Count diff types, this should be 3 insertions and deletions from these manifests
	insertionCount, deletionCount, equalCount := countDiffTypes(diffs)
	expectedCountIns, expectedCountDel, expectedCountEq := 11, 0, 0

	require.Equal(t, expectedCountIns, insertionCount)
	require.Equal(t, expectedCountDel, deletionCount)
	require.Equal(t, expectedCountEq, equalCount)
}
