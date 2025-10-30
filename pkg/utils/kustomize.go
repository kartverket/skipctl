//revive:disable-next-line var-naming
package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kartverket/skipctl/pkg/git"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type pathCollector struct {
	files           []string
	visited         map[string]bool
	absKustomizeDir string
}

func (c *pathCollector) addPath(relPath string) error {
	if strings.HasPrefix(relPath, "http://") || strings.HasPrefix(relPath, "https://") {
		return nil
	}

	absPath, err := filepath.Abs(filepath.Join(c.absKustomizeDir, relPath))

	if err != nil {
		return err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return c.addDirectory(absPath, relPath)
	}

	c.files = append(c.files, absPath)
	return nil
}

func (c *pathCollector) addDirectory(absPath, relPath string) error {
	kustPath := filepath.Join(absPath, "kustomization.yaml")

	if _, err := os.Stat(kustPath); err == nil {
		nestedContent, readErr := os.ReadFile(kustPath)
		if readErr != nil {
			return readErr
		}
		nestedFiles, collectErr := CollectKustomizationFiles(relPath, nestedContent, c.absKustomizeDir, c.visited)
		if collectErr != nil {
			return collectErr
		}
		c.files = append(c.files, nestedFiles...)
		return nil
	}

	return filepath.WalkDir(absPath, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		c.files = append(c.files, path)
		return nil
	})
}

func (c *pathCollector) addPaths(paths []string) error {
	for _, path := range paths {
		if err := c.addPath(path); err != nil {
			return err
		}
	}
	return nil
}

//nolint:staticcheck,gocognit, gocyclo, cyclop, funlen // reason: legacy API use and complex control flow;
func CollectKustomizationFiles(kustomizeDir string, kustContent []byte, baseDir string, visited map[string]bool) ([]string, error) {
	if visited == nil {
		visited = make(map[string]bool)
	}
	var err error

	absKustomizeDir, err := filepath.Abs(filepath.Join(baseDir, kustomizeDir))
	if err != nil {
		return nil, err
	}

	if visited[absKustomizeDir] {
		return nil, nil
	}
	visited[absKustomizeDir] = true

	kust := &types.Kustomization{}
	if yamlErr := yaml.Unmarshal(kustContent, kust); yamlErr != nil {
		return nil, fmt.Errorf("parse kustomization.yaml: %w", yamlErr)
	}

	collector := &pathCollector{
		files:           []string{filepath.Join(absKustomizeDir, "kustomization.yaml")},
		visited:         visited,
		absKustomizeDir: absKustomizeDir,
	}

	if err = collector.addPaths(kust.Resources); err != nil {
		return nil, err
	}
	if err = collector.addPaths(kust.Bases); err != nil {
		return nil, err
	}
	if err = collector.addPaths(kust.Components); err != nil {
		return nil, err
	}
	if err = collector.addPaths(kust.Crds); err != nil {
		return nil, err
	}
	if err = collector.addPaths(kust.Configurations); err != nil {
		return nil, err
	}
	if err = collector.addPaths(kust.Generators); err != nil {
		return nil, err
	}
	if err = collector.addPaths(kust.Transformers); err != nil {
		return nil, err
	}
	if err = collector.addPaths(kust.Validators); err != nil {
		return nil, err
	}

	for _, openAPIPath := range kust.OpenAPI {
		if err = collector.addPath(openAPIPath); err != nil {
			return nil, err
		}
	}

	for _, cm := range kust.ConfigMapGenerator {
		if err = collector.addPaths(cm.FileSources); err != nil {
			return nil, err
		}
		if err = collector.addPaths(cm.EnvSources); err != nil {
			return nil, err
		}
	}

	for _, secret := range kust.SecretGenerator {
		if err = collector.addPaths(secret.FileSources); err != nil {
			return nil, err
		}
		if err = collector.addPaths(secret.EnvSources); err != nil {
			return nil, err
		}
	}

	for _, patch := range kust.Patches {
		if patch.Path != "" {
			if err = collector.addPath(patch.Path); err != nil {
				return nil, err
			}
		}
	}

	for _, patch := range kust.PatchesStrategicMerge {
		if err = collector.addPath(string(patch)); err != nil {
			return nil, err
		}
	}

	for _, patch := range kust.PatchesJson6902 {
		if patch.Path != "" {
			if err = collector.addPath(patch.Path); err != nil {
				return nil, err
			}
		}
	}

	for _, replacement := range kust.Replacements {
		if replacement.Path != "" {
			if err = collector.addPath(replacement.Path); err != nil {
				return nil, err
			}
		}
	}

	return collector.files, nil
}

func CopyKustomziationFilesToTmpDirAtGitRef(filesToCopy []string, ref string) (string, error) {
	tmpDir, dirErr := os.MkdirTemp(os.TempDir(), "kustomize-diff-tmp")

	if dirErr != nil {
		return "", dirErr
	}

	// Copy all collected files to temp directory maintaining structure
	for _, filePath := range filesToCopy {
		content, gitErr := git.GetFileContentAtRef(filePath, ref)
		if gitErr != nil {
			continue
		}
		tmpFilePath := filepath.Join(tmpDir, filePath)

		if mkdirErr := os.MkdirAll(filepath.Dir(tmpFilePath), 0755); mkdirErr != nil {
			return "", mkdirErr
		}
		if writeErr := os.WriteFile(tmpFilePath, []byte(*content), 0600); writeErr != nil {
			return "", writeErr
		}
	}

	return tmpDir, nil
}
