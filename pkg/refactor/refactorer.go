package refactor

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	api "github.com/kartverket/skipctl/pkg/api/v1"
	"github.com/kartverket/skipctl/pkg/auth"
	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/kartverket/skipctl/pkg/prompts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const chunkSize = 64 * 1024 // 64KB chunks

// Token limits for Vertex AI models (approximate, accounting for prompt + response)
const (
	maxPromptSizeBytes     = 1 * 1024 * 1024 // 1MB - conservative limit for most models
	warnPromptSizeBytes    = 500 * 1024      // 500KB - warning threshold
	maxPromptSizeReadable  = "1MB"
	warnPromptSizeReadable = "500KB"
)

// Manifest sends manifest files to the server for AI refactoring via gRPC
func Manifest(ctx context.Context, docs []*manifest.Document, serverAddr string) error {
	if len(docs) == 0 {
		return errors.New("no documents provided for refactoring")
	}

	// Get authentication credentials
	perRPCCreds, err := auth.NewADCBackedRPCCredentials()
	if err != nil {
		return fmt.Errorf("failed to create authentication credentials: %w", err)
	}

	// Use TLS with system's root CA certificates for remote servers
	// Use insecure credentials only for localhost
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithPerRPCCredentials(perRPCCreds))

	if strings.HasPrefix(serverAddr, "localhost:") || strings.HasPrefix(serverAddr, "127.0.0.1:") {
		// For localhost, check if TLS is available by trying with InsecureSkipVerify first
		// #nosec G402 - InsecureSkipVerify is acceptable for localhost development with self-signed certificates
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}
		tlsCreds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(tlsCreds))
		slog.InfoContext(ctx, "Using TLS with self-signed certificate support for localhost")
	} else {
		// Use TLS for remote servers with proper certificate validation
		tlsCreds := credentials.NewTLS(nil)
		opts = append(opts, grpc.WithTransportCredentials(tlsCreds))
		slog.InfoContext(ctx, "Using TLS connection", "server", serverAddr)
	}

	// Connect to the server
	conn, err := grpc.NewClient(serverAddr, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close()

	client := api.NewAIServiceClient(conn)

	// Extract and append imported files
	docs, err = appendImportedFiles(docs)
	if err != nil {
		slog.WarnContext(ctx, "Failed to extract imports", "error", err.Error())
	}

	// Build combined content from all documents
	combinedContent := buildCombinedContent(ctx, docs)
	contentBytes := []byte(combinedContent)

	// Create the prompt with system instruction
	prompt := prompts.RefactorSystemPrompt + "\n\n" + "Please refactor the libsonnet file in the /application, and use the other files for context:\n\n" + combinedContent

	// Validate prompt size before sending
	if err = validatePromptSize(prompt); err != nil {
		return err
	}

	// Stream content to server
	resp, err := streamToServer(ctx, client, docs[0], contentBytes, prompt)
	if err != nil {
		return err
	}

	// Write the refactored content to file
	return writeRefactoredOutput(ctx, resp)
}

func validatePromptSize(prompt string) error {
	promptSize := len(prompt)

	if promptSize > maxPromptSizeBytes {
		return fmt.Errorf("prompt size (%d bytes) exceeds maximum allowed size (%s). "+
			"Consider reducing the number of files or file sizes",
			promptSize, maxPromptSizeReadable)
	}

	if promptSize > warnPromptSizeBytes {
		fmt.Fprintf(os.Stderr, "Warning: prompt size is large (%d bytes, over %s). "+
			"This may approach model token limits and could fail or be truncated.\n",
			promptSize, warnPromptSizeReadable)
	}

	return nil
}

func appendImportedFiles(docs []*manifest.Document) ([]*manifest.Document, error) {
	mainDoc := docs[0]
	importedDocs, err := extractImportedFiles(mainDoc)
	if err != nil {
		return docs, err
	}
	slog.Info("Discovered imports", "count", len(importedDocs), "mainFile", mainDoc.Name)
	for _, doc := range importedDocs {
		slog.Info("  Imported file", "name", doc.Name, "path", doc.Path)
	}
	return append(docs, importedDocs...), nil
}

func buildCombinedContent(ctx context.Context, docs []*manifest.Document) string {
	var combinedContent strings.Builder
	for i, doc := range docs {
		if i > 0 {
			combinedContent.WriteString("\n\n---\n\n")
		}
		combinedContent.WriteString(fmt.Sprintf("File: %s\n\n", doc.Name))
		combinedContent.WriteString(doc.Content)

		// If it's a Jsonnet file, render it to JSON and include the output
		if isJsonnetFile(doc.Extension) && !doc.Rendered {
			appendRenderedContent(ctx, &combinedContent, doc)
		}
	}
	return combinedContent.String()
}

func appendRenderedContent(ctx context.Context, builder *strings.Builder, doc *manifest.Document) {
	slog.InfoContext(ctx, "Attempting to render Jsonnet file", "file", doc.Name, "extension", doc.Extension)
	renderedContent, err := renderDocument(doc)
	if err != nil {
		slog.WarnContext(ctx, "Failed to render Jsonnet file", "file", doc.Name, "error", err.Error())
		fmt.Fprintf(os.Stderr, "Warning: failed to render %s: %v\n", doc.Path, err)
		return
	}
	slog.InfoContext(ctx, "Successfully rendered Jsonnet file", "file", doc.Name, "size", len(renderedContent))
	builder.WriteString("\n\n---\n\nRendered JSON output:\n")
	builder.WriteString(renderedContent)
}

func streamToServer(ctx context.Context, client api.AIServiceClient, firstDoc *manifest.Document, contentBytes []byte, prompt string) (*api.RefactorToArgokitv2Response, error) {
	stream, err := client.RefactorToArgokitv2(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}

	// Send file in chunks
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

		if sendErr := stream.Send(req); sendErr != nil {
			return nil, fmt.Errorf("failed to send chunk: %w", sendErr)
		}
	}

	// Close and receive response
	resp, err := stream.CloseAndRecv()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, errors.New("unexpected EOF from server")
		}
		return nil, fmt.Errorf("failed to receive response: %w", err)
	}

	return resp, nil
}

func generateUniqueOutputPath(baseName, ext string) (string, error) {
	// Start with the base filename and, if it exists, append a numeric suffix.
	newPath := fmt.Sprintf("%s%s", baseName, ext)
	if _, err := os.Stat(newPath); errors.Is(err, os.ErrNotExist) {
		return newPath, nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("failed to stat output file %q: %w", newPath, err)
	}

	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s_%d%s", baseName, i, ext)
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("failed to stat output file %q: %w", candidate, err)
		}
	}
}

func writeRefactoredOutput(ctx context.Context, resp *api.RefactorToArgokitv2Response) error {
	if resp.GetResponse() == "" {
		return errors.New("empty response from server")
	}

	newPath, err := generateUniqueOutputPath("vertexAI_output", ".libsonnet")
	if err != nil {
		return fmt.Errorf("failed to determine output file path: %w", err)
	}
	if err = os.WriteFile(newPath, []byte(resp.GetResponse()), 0600); err != nil {
		return fmt.Errorf("failed to write refactored content to file: %w", err)
	}

	slog.InfoContext(ctx, "Successfully refactored", "output", newPath)
	return nil
}

func extractImportedFiles(doc *manifest.Document) ([]*manifest.Document, error) {
	// Get absolute path to the directory containing the main document
	absDocPath, err := filepath.Abs(doc.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", doc.Path, err)
	}

	// Track seen files to avoid duplicates and circular imports
	seen := make(map[string]bool)
	var importedDocs []*manifest.Document

	// Use a queue for breadth-first traversal of imports
	toProcess := []*manifest.Document{doc}
	seen[filepath.Clean(absDocPath)] = true

	for len(toProcess) > 0 {
		current := toProcess[0]
		toProcess = toProcess[1:]

		// Extract imports from current document
		imports := extractImportsFromDocument(current, seen)

		// Add newly discovered imports to the result and queue for processing
		for _, imp := range imports {
			importedDocs = append(importedDocs, imp)
			toProcess = append(toProcess, imp)
		}
	}

	return importedDocs, nil
}

func extractImportsFromDocument(doc *manifest.Document, seen map[string]bool) []*manifest.Document {
	// Matches: import 'file.libsonnet', importstr 'file.txt', import "file.libsonnet"
	importRegex := regexp.MustCompile(`(?:import|importstr)\s+['"]([^'"]+)['"]`)

	matches := importRegex.FindAllStringSubmatch(doc.Content, -1)
	if len(matches) == 0 {
		return nil
	}

	var importedDocs []*manifest.Document

	const minMatchLength = 2
	for _, match := range matches {
		if len(match) < minMatchLength {
			continue
		}

		importPath := match[1]

		// Resolve relative path from the current document's directory
		currentDocDir := filepath.Dir(doc.Path)
		absPath := filepath.Join(currentDocDir, importPath)

		// Clean the path to resolve . and .. properly
		absPath = filepath.Clean(absPath)

		// Avoid duplicates and circular imports
		if seen[absPath] {
			continue
		}
		seen[absPath] = true

		// Read the imported file using the absolute path
		content, readErr := os.ReadFile(absPath)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to read imported file %s: %v\n", absPath, readErr)
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

	return importedDocs
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

	ctx := context.Background()
	slog.InfoContext(ctx, "Rendered content", "content", docCopy.Content)

	// Return the rendered content
	return docCopy.Content, nil
}
