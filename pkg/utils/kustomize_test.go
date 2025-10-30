package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectKustomizationFiles(t *testing.T) {
	tmpDir := t.TempDir()

	baseDir := filepath.Join(tmpDir, "base")
	overlayDir := filepath.Join(tmpDir, "overlays", "prod")
	os.MkdirAll(baseDir, 0755)
	os.MkdirAll(overlayDir, 0755)

	baseDeployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 1`
	os.WriteFile(filepath.Join(baseDir, "deployment.yaml"), []byte(baseDeployment), 0644)

	baseKustomization := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - deployment.yaml`
	os.WriteFile(filepath.Join(baseDir, "kustomization.yaml"), []byte(baseKustomization), 0644)

	configData := "ENV=production"
	os.WriteFile(filepath.Join(overlayDir, "config.env"), []byte(configData), 0644)

	patch := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 3`
	os.WriteFile(filepath.Join(overlayDir, "replica-patch.yaml"), []byte(patch), 0644)

	overlayKustomization := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ../../base
patches:
  - path: replica-patch.yaml
configMapGenerator:
  - name: app-config
    envs:
      - config.env`
	os.WriteFile(filepath.Join(overlayDir, "kustomization.yaml"), []byte(overlayKustomization), 0644)

	files, err := CollectKustomizationFiles(".", []byte(overlayKustomization), overlayDir, nil)
	if err != nil {
		t.Fatalf("CollectKustomizationFiles failed: %v", err)
	}

	expectedFiles := []string{
		filepath.Join(overlayDir, "kustomization.yaml"),
		filepath.Join(baseDir, "kustomization.yaml"),
		filepath.Join(baseDir, "deployment.yaml"),
		filepath.Join(overlayDir, "replica-patch.yaml"),
		filepath.Join(overlayDir, "config.env"),
	}

	if len(files) != len(expectedFiles) {
		t.Logf("Got files: %v", files)
		t.Errorf("Expected %d files, got %d", len(expectedFiles), len(files))
	}

	fileMap := make(map[string]bool)
	for _, f := range files {
		fileMap[f] = true
	}

	for _, expected := range expectedFiles {
		if !fileMap[expected] {
			t.Errorf("Expected file not found: %s", expected)
		}
	}
}
