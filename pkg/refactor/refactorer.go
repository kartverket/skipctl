package refactor

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	api "github.com/kartverket/skipctl/pkg/api/v1"
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

	// Combine all document contents as context
	var combinedContent strings.Builder
	for i, doc := range docs {
		if i > 0 {
			combinedContent.WriteString("\n\n---\n\n")
		}
		combinedContent.WriteString(fmt.Sprintf("File: %s\n\n", doc.Name))
		combinedContent.WriteString(doc.Content)

		// If it's a Jsonnet file, render it to JSON and include the output
		if isJsonnetFile(doc.Path) && !doc.Rendered {
			renderedContent, err := renderDocument(ctx, doc)
			if err != nil {
				// Log the error but continue - we'll still have the original Jsonnet
				fmt.Fprintf(os.Stderr, "Warning: failed to render %s: %v\n", doc.Path, err)
			} else {
				combinedContent.WriteString("\n\nRendered JSON output:\n")
				combinedContent.WriteString(renderedContent)
			}
		}
	}

	contentBytes := []byte(combinedContent.String())

	// Create the prompt with system instruction
	prompt := prompts.RefactorSystemPrompt + "\n\n" + "Please refactor the following manifest files:\n\n" + combinedContent.String()

	// Open stream
	stream, err := client.AnalyzeFile(ctx)
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

		req := &api.AnalyzeFileRequest{
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
		if err == io.EOF {
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

func isJsonnetFile(path string) bool {
	lowerPath := strings.ToLower(path)
	return strings.HasSuffix(lowerPath, ".jsonnet") || strings.HasSuffix(lowerPath, ".libsonnet")
}

func renderDocument(ctx context.Context, doc *manifest.Document) (string, error) {
	// Create a copy of the document for rendering
	docCopy := &manifest.Document{
		Name:      doc.Name,
		Content:   doc.Content,
		Extension: doc.Extension,
		Path:      doc.Path,
		Rendered:  doc.Rendered,
	}

	// Create a discard logger to suppress render output
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	renderer := manifest.NewRenderer(discardLogger)

	// Render the document in-place
	err := renderer.Render(docCopy)
	if err != nil {
		return "", fmt.Errorf("failed to render: %w", err)
	}

	// Return the rendered content
	return docCopy.Content, nil
}
