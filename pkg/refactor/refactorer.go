package refactor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	api "github.com/kartverket/skipctl/pkg/api/v1"
	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/kartverket/skipctl/pkg/prompts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const chunkSize = 64 * 1024 // 64KB chunks

// RefactorManifest sends manifest files to the server for AI refactoring via gRPC
func RefactorManifest(ctx context.Context, docs []*manifest.Document, serverAddr string) error {
	if len(docs) == 0 {
		return fmt.Errorf("no documents provided for refactoring")
	}

	// Connect to the server
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close()

	client := api.NewAIServiceClient(conn)

	mainDoc := docs[0]

	// Extract imported files
	importedDocs, err := extractImportedFiles(mainDoc)
	if err != nil {
		slog.WarnContext(ctx, "Failed to extract imports", "error", err.Error())
	} else {
		docs = append(docs, importedDocs...)
	}

	// Combine all document contents as context
	var combinedContent strings.Builder
	for i, doc := range docs {
		if i > 0 {
			combinedContent.WriteString("\n\n---\n\n")
		}
		combinedContent.WriteString(fmt.Sprintf("File: %s\n\n", doc.Name))
		combinedContent.WriteString(doc.Content)

		// If it's a Jsonnet file, render it to JSON and include the output
		if isJsonnetFile(doc.Extension) && !doc.Rendered {
			slog.InfoContext(ctx, "Attempting to render Jsonnet file", "file", doc.Name, "extension", doc.Extension)
			renderedContent, err := renderDocument(doc)
			if err != nil {
				// Log the error but continue - we'll still have the original Jsonnet
				slog.WarnContext(ctx, "Failed to render Jsonnet file", "file", doc.Name, "error", err.Error())
				fmt.Fprintf(os.Stderr, "Warning: failed to render %s: %v\n", doc.Path, err)
			} else {
				slog.InfoContext(ctx, "Successfully rendered Jsonnet file", "file", doc.Name, "size", len(renderedContent))
				combinedContent.WriteString("\n\n---\n\nRendered JSON output:\n")
				combinedContent.WriteString(renderedContent)
			}
		}
	}

	contentBytes := []byte(combinedContent.String())

	// Create the prompt with system instruction
	prompt := prompts.RefactorSystemPrompt + "\n\n" + "Please refactor the manifest file in the /application, and use the other files for context:\n\n" + combinedContent.String()

	// Open stream
	stream, err := client.RefactorToArgokitv2(ctx)
	if err != nil {
		return fmt.Errorf("failed to open stream: %w", err)
	}

	// Send file in chunks
	firstDoc := docs[0]
	for offset := 0; offset < len(contentBytes); offset += chunkSize {
		end := offset + chunkSize
		if end > len(contentBytes) {
			end = len(contentBytes)
		}

		req := &api.RefactorToArgokitv2Request{
			Chunk: contentBytes[offset:end],
		}

		// Send metadata in first chunk
		if offset == 0 {
			req.FileName = firstDoc.Name
			req.MimeType = "text/plain"
			req.Prompt = prompt
		}

		if err := stream.Send(req); err != nil {
			return fmt.Errorf("failed to send chunk: %w", err)
		}
	}

	// Close and receive response
	resp, err := stream.CloseAndRecv()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("unexpected EOF from server")
		}
		return fmt.Errorf("failed to receive response: %w", err)
	}

	// Write the refactored content to file
	if resp.Response != "" {
		nameWithoutExt := strings.TrimSuffix(firstDoc.Name, firstDoc.Extension)
		newPath := fmt.Sprintf("%s.refactored.jsonnet", nameWithoutExt)

		err := os.WriteFile(newPath, []byte(resp.Response), 0644)
		if err != nil {
			return fmt.Errorf("failed to write refactored content to file: %w", err)
		}

		fmt.Printf("Successfully refactored to: %s\n", newPath)
		return nil
	}

	return fmt.Errorf("empty response from server")
}

func extractImportedFiles(doc *manifest.Document) ([]*manifest.Document, error) {
	{
		// Matches: import 'file.libsonnet', importstr 'file.txt', import "file.libsonnet"
		importRegex := regexp.MustCompile(`(?:import|importstr)\s+['"]([^'"]+)['"]`)

		matches := importRegex.FindAllStringSubmatch(doc.Content, -1)
		if len(matches) == 0 {
			return nil, nil
		}

		var importedDocs []*manifest.Document
		baseDir := filepath.Dir(doc.Path)
		seen := make(map[string]bool)

		for _, match := range matches {
			if len(match) < 2 {
				continue
			}

			importPath := match[1]

			// Resolve relative path
			absPath := filepath.Join(baseDir, importPath)

			// Avoid duplicates
			if seen[absPath] {
				continue
			}
			seen[absPath] = true

			// Read the imported file
			content, err := os.ReadFile(absPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to read imported file %s: %v\n", absPath, err)
				continue
			}

			importedDoc := &manifest.Document{
				Name:      filepath.Base(absPath),
				Content:   string(content),
				Extension: filepath.Ext(absPath),
				Path:      absPath,
				Rendered:  false,
			}

			importedDocs = append(importedDocs, importedDoc)
		}

		return importedDocs, nil
	}
}

func isJsonnetFile(path string) bool {
	lowerPath := strings.ToLower(path)
	return strings.HasSuffix(lowerPath, ".jsonnet") || strings.HasSuffix(lowerPath, ".libsonnet")
}

func renderDocument(doc *manifest.Document) (string, error) {
	// Create a copy of the document for rendering
	docCopy := &manifest.Document{
		Name:      doc.Name,
		Content:   doc.Content,
		Extension: doc.Extension,
		Path:      doc.Path,
		Rendered:  doc.Rendered,
	}

	renderer := manifest.NewRenderer(logging.Logger())

	// Render the document in-place
	err := renderer.Render(docCopy)
	if err != nil {
		return "", fmt.Errorf("failed to render: %w", err)
	}

	slog.Info("Rendered content", "content", docCopy.Content)

	// Return the rendered content
	return docCopy.Content, nil
}
