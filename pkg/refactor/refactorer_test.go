package refactor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kartverket/skipctl/pkg/manifest"
)

func TestValidatePromptSize(t *testing.T) {
	tests := []struct {
		name        string
		promptSize  int
		expectError bool
		expectWarn  bool
	}{
		{
			name:        "small prompt - no warning",
			promptSize:  100 * 1024, // 100KB
			expectError: false,
			expectWarn:  false,
		},
		{
			name:        "medium prompt - warning threshold",
			promptSize:  600 * 1024, // 600KB
			expectError: false,
			expectWarn:  true,
		},
		{
			name:        "large prompt - exceeds max",
			promptSize:  2 * 1024 * 1024, // 2MB
			expectError: true,
			expectWarn:  false,
		},
		{
			name:        "exactly at max - allowed",
			promptSize:  maxPromptSizeBytes,
			expectError: false,
			expectWarn:  true,
		},
		{
			name:        "just over max - error",
			promptSize:  maxPromptSizeBytes + 1,
			expectError: true,
			expectWarn:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a prompt of the specified size
			prompt := strings.Repeat("a", tt.promptSize)

			err := validatePromptSize(prompt)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestExtractImportedFiles(t *testing.T) {
	// Create a temporary directory with test files
	tempDir := t.TempDir()

	// Create test files
	mainContent := `
local lib1 = import 'lib1.libsonnet';
local lib2 = import "lib2.libsonnet";

{
  main: lib1.value + lib2.value,
}
`
	lib1Content := `
local helper = import 'helper.libsonnet';
{
  value: helper.getValue(),
}
`
	lib2Content := `
{
  value: 'from lib2',
}
`
	helperContent := `
{
  getValue(): 'from helper',
}
`

	// Write test files
	mainPath := filepath.Join(tempDir, "main.jsonnet")
	lib1Path := filepath.Join(tempDir, "lib1.libsonnet")
	lib2Path := filepath.Join(tempDir, "lib2.libsonnet")
	helperPath := filepath.Join(tempDir, "helper.libsonnet")

	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main file: %v", err)
	}
	if err := os.WriteFile(lib1Path, []byte(lib1Content), 0644); err != nil {
		t.Fatalf("failed to write lib1 file: %v", err)
	}
	if err := os.WriteFile(lib2Path, []byte(lib2Content), 0644); err != nil {
		t.Fatalf("failed to write lib2 file: %v", err)
	}
	if err := os.WriteFile(helperPath, []byte(helperContent), 0644); err != nil {
		t.Fatalf("failed to write helper file: %v", err)
	}

	// Create main document
	mainDoc := &manifest.Document{
		Name:      "main.jsonnet",
		Content:   mainContent,
		Extension: ".jsonnet",
		Path:      mainPath,
	}

	// Extract imported files
	importedDocs, err := extractImportedFiles(mainDoc)
	if err != nil {
		t.Fatalf("extractImportedFiles failed: %v", err)
	}

	// Verify we found all imports
	if len(importedDocs) != 3 {
		t.Errorf("expected 3 imported files, got %d", len(importedDocs))
	}

	// Verify the imported file names
	foundFiles := make(map[string]bool)
	for _, doc := range importedDocs {
		foundFiles[doc.Name] = true
	}

	expectedFiles := []string{"lib1.libsonnet", "lib2.libsonnet", "helper.libsonnet"}
	for _, expected := range expectedFiles {
		if !foundFiles[expected] {
			t.Errorf("expected to find imported file %s", expected)
		}
	}
}

func TestExtractImportedFiles_CircularImports(t *testing.T) {
	tempDir := t.TempDir()

	// Create files with circular imports
	file1Content := `
local file2 = import 'file2.libsonnet';
{ value: 'file1' }
`
	file2Content := `
local file1 = import 'file1.libsonnet';
{ value: 'file2' }
`

	file1Path := filepath.Join(tempDir, "file1.libsonnet")
	file2Path := filepath.Join(tempDir, "file2.libsonnet")

	if err := os.WriteFile(file1Path, []byte(file1Content), 0644); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}
	if err := os.WriteFile(file2Path, []byte(file2Content), 0644); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}

	mainDoc := &manifest.Document{
		Name:      "file1.libsonnet",
		Content:   file1Content,
		Extension: ".libsonnet",
		Path:      file1Path,
	}

	// Should handle circular imports without infinite loop
	importedDocs, err := extractImportedFiles(mainDoc)
	if err != nil {
		t.Fatalf("extractImportedFiles failed: %v", err)
	}

	// Should only return file2 once (not enter infinite loop)
	if len(importedDocs) != 1 {
		t.Errorf("expected 1 imported file, got %d", len(importedDocs))
	}
	if importedDocs[0].Name != "file2.libsonnet" {
		t.Errorf("expected file2.libsonnet, got %s", importedDocs[0].Name)
	}
}

func TestExtractImportedFiles_MissingFile(t *testing.T) {
	tempDir := t.TempDir()

	mainContent := `
local missing = import 'nonexistent.libsonnet';
{ value: 'test' }
`
	mainPath := filepath.Join(tempDir, "main.jsonnet")
	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main file: %v", err)
	}

	mainDoc := &manifest.Document{
		Name:      "main.jsonnet",
		Content:   mainContent,
		Extension: ".jsonnet",
		Path:      mainPath,
	}

	// Should not error, but skip missing file
	importedDocs, err := extractImportedFiles(mainDoc)
	if err != nil {
		t.Fatalf("extractImportedFiles failed: %v", err)
	}

	if len(importedDocs) != 0 {
		t.Errorf("expected 0 imported files (missing file should be skipped), got %d", len(importedDocs))
	}
}

func TestBuildCombinedContent(t *testing.T) {
	docs := []*manifest.Document{
		{
			Name:      "main.jsonnet",
			Content:   "{ main: 'content' }",
			Extension: ".jsonnet",
			Path:      "/path/to/main.jsonnet",
			Rendered:  true, // Already rendered, skip rendering
		},
		{
			Name:      "lib.libsonnet",
			Content:   "{ lib: 'helper' }",
			Extension: ".libsonnet",
			Path:      "/path/to/lib.libsonnet",
			Rendered:  true,
		},
	}

	ctx := context.Background()
	combined := buildCombinedContent(ctx, docs)

	// Check that both files are included
	if !strings.Contains(combined, "File: main.jsonnet") {
		t.Error("combined content should contain main.jsonnet")
	}
	if !strings.Contains(combined, "File: lib.libsonnet") {
		t.Error("combined content should contain lib.libsonnet")
	}
	if !strings.Contains(combined, "{ main: 'content' }") {
		t.Error("combined content should contain main content")
	}
	if !strings.Contains(combined, "{ lib: 'helper' }") {
		t.Error("combined content should contain lib content")
	}

	// Check separator
	if !strings.Contains(combined, "\n\n---\n\n") {
		t.Error("combined content should contain separator")
	}
}

func TestIsJsonnetFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{".jsonnet", true},
		{".libsonnet", true},
		{".JSONNET", true},
		{".LibSonnet", true},
		{".json", false},
		{".yaml", false},
		{".txt", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isJsonnetFile(tt.path)
			if result != tt.expected {
				t.Errorf("isJsonnetFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestGenerateUniqueOutputPath(t *testing.T) {
	tempDir := t.TempDir()

	t.Chdir(tempDir)

	t.Run("no existing file", func(t *testing.T) {
		path, pathErr := generateUniqueOutputPath("test", ".txt")
		if pathErr != nil {
			t.Fatalf("unexpected error: %v", pathErr)
		}
		if path != "test.txt" {
			t.Errorf("expected test.txt, got %s", path)
		}
	})

	t.Run("existing file - generate unique name", func(t *testing.T) {
		// Create test.txt
		if writeErr := os.WriteFile("output.txt", []byte("test"), 0644); writeErr != nil {
			t.Fatalf("failed to create test file: %v", writeErr)
		}

		path, pathErr := generateUniqueOutputPath("output", ".txt")
		if pathErr != nil {
			t.Fatalf("unexpected error: %v", pathErr)
		}
		if path != "output_1.txt" {
			t.Errorf("expected output_1.txt, got %s", path)
		}
	})

	t.Run("multiple existing files", func(t *testing.T) {
		// Create multi.txt and multi_1.txt
		if writeErr := os.WriteFile("multi.txt", []byte("test"), 0644); writeErr != nil {
			t.Fatalf("failed to create file: %v", writeErr)
		}
		if writeErr := os.WriteFile("multi_1.txt", []byte("test"), 0644); writeErr != nil {
			t.Fatalf("failed to create file: %v", writeErr)
		}

		path, pathErr := generateUniqueOutputPath("multi", ".txt")
		if pathErr != nil {
			t.Fatalf("unexpected error: %v", pathErr)
		}
		if path != "multi_2.txt" {
			t.Errorf("expected multi_2.txt, got %s", path)
		}
	})
}

func TestAppendImportedFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create a simple main file with one import
	mainContent := `local lib = import 'lib.libsonnet';`
	libContent := `{ value: 'test' }`

	mainPath := filepath.Join(tempDir, "main.jsonnet")
	libPath := filepath.Join(tempDir, "lib.libsonnet")

	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main file: %v", err)
	}
	if err := os.WriteFile(libPath, []byte(libContent), 0644); err != nil {
		t.Fatalf("failed to write lib file: %v", err)
	}

	mainDoc := &manifest.Document{
		Name:      "main.jsonnet",
		Content:   mainContent,
		Extension: ".jsonnet",
		Path:      mainPath,
	}

	docs := []*manifest.Document{mainDoc}

	result, err := appendImportedFiles(docs)
	if err != nil {
		t.Fatalf("appendImportedFiles failed: %v", err)
	}

	// Should have original + imported file
	if len(result) != 2 {
		t.Errorf("expected 2 documents (main + lib), got %d", len(result))
	}

	// First should be main
	if result[0].Name != "main.jsonnet" {
		t.Errorf("expected first doc to be main.jsonnet, got %s", result[0].Name)
	}

	// Second should be lib
	if result[1].Name != "lib.libsonnet" {
		t.Errorf("expected second doc to be lib.libsonnet, got %s", result[1].Name)
	}
}

//nolint:gocognit // Test function complexity is acceptable for comprehensive test coverage
func TestExtractImportsFromDocument(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedCount int
		expectedNames []string
	}{
		{
			name:          "no imports",
			content:       `{ value: 'test' }`,
			expectedCount: 0,
		},
		{
			name: "single quote import",
			content: `
				local lib = import 'lib.libsonnet';
				{ value: lib.value }
			`,
			expectedCount: 1,
			expectedNames: []string{"lib.libsonnet"},
		},
		{
			name: "double quote import",
			content: `
				local lib = import "lib.libsonnet";
				{ value: lib.value }
			`,
			expectedCount: 1,
			expectedNames: []string{"lib.libsonnet"},
		},
		{
			name: "importstr",
			content: `
				local data = importstr 'data.txt';
				{ value: data }
			`,
			expectedCount: 1,
			expectedNames: []string{"data.txt"},
		},
		{
			name: "multiple imports",
			content: `
				local lib1 = import 'lib1.libsonnet';
				local lib2 = import "lib2.libsonnet";
				local data = importstr 'data.txt';
			`,
			expectedCount: 3,
			expectedNames: []string{"lib1.libsonnet", "lib2.libsonnet", "data.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Create the files that will be imported
			for _, name := range tt.expectedNames {
				filePath := filepath.Join(tempDir, name)
				if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
					t.Fatalf("failed to create %s: %v", name, err)
				}
			}

			// Create main document
			mainPath := filepath.Join(tempDir, "main.jsonnet")
			if err := os.WriteFile(mainPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write main file: %v", err)
			}

			doc := &manifest.Document{
				Name:      "main.jsonnet",
				Content:   tt.content,
				Extension: ".jsonnet",
				Path:      mainPath,
			}

			seen := make(map[string]bool)
			seen[filepath.Clean(mainPath)] = true

			imports := extractImportsFromDocument(doc, seen)

			if len(imports) != tt.expectedCount {
				t.Errorf("expected %d imports, got %d", tt.expectedCount, len(imports))
			}

			// Verify expected names
			for _, expectedName := range tt.expectedNames {
				found := false
				for _, imp := range imports {
					if imp.Name == expectedName {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected to find import %s", expectedName)
				}
			}
		})
	}
}
