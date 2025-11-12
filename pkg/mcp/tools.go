package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/manifest"
)

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

func (s *Server) handleToolsList(req *JSONRPCRequest) *JSONRPCResponse {
	tools := []Tool{
		{
			Name:        "render_manifest",
			Description: "Render a manifest file (Jsonnet/YAML/Kustomize) to see the final YAML output. Use this to preview what will be deployed.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file": map[string]interface{}{
						"type":        "string",
						"description": "Path to manifest file (relative or absolute)",
					},
				},
				"required": []string{"file"},
			},
		},
		{
			Name:        "diff_manifest",
			Description: "Compare manifest changes between current state and a git reference. Shows what will change if the manifest is applied.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file": map[string]interface{}{
						"type":        "string",
						"description": "Path to manifest file",
					},
					"ref": map[string]interface{}{
						"type":        "string",
						"description": "Git reference to compare against (default: HEAD)",
					},
				},
				"required": []string{"file"},
			},
		},
		{
			Name:        "validate_manifest",
			Description: "Validate manifest syntax and structure. Checks if the file can be parsed and rendered.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file": map[string]interface{}{
						"type":        "string",
						"description": "Path to manifest file",
					},
				},
				"required": []string{"file"},
			},
		},
		{
			Name:        "format_manifest",
			Description: "Format a manifest file according to standard conventions (YAML/Jsonnet).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file": map[string]interface{}{
						"type":        "string",
						"description": "Path to manifest file",
					},
				},
				"required": []string{"file"},
			},
		},
		{
			Name:        "list_manifests",
			Description: "List all manifest files in the current directory or workspace.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Directory path to search (default: current directory)",
					},
				},
			},
		},
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Result: ToolsListResult{
			Tools: tools,
		},
		ID: req.ID,
	}
}

type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ToolCallResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (s *Server) handleToolsCall(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    -32602,
				Message: fmt.Sprintf("Invalid params: %v", err),
			},
			ID: req.ID,
		}
	}

	var result ToolCallResult
	var err error

	switch params.Name {
	case "render_manifest":
		result, err = s.renderManifest(ctx, params.Arguments)
	case "diff_manifest":
		result, err = s.diffManifest(ctx, params.Arguments)
	case "validate_manifest":
		result, err = s.validateManifest(ctx, params.Arguments)
	case "format_manifest":
		result, err = s.formatManifest(ctx, params.Arguments)
	case "list_manifests":
		result, err = s.listManifests(ctx, params.Arguments)
	default:
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    -32601,
				Message: fmt.Sprintf("Unknown tool: %s", params.Name),
			},
			ID: req.ID,
		}
	}

	if err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Result: ToolCallResult{
				Content: []ContentBlock{
					{
						Type: "text",
						Text: fmt.Sprintf("Error: %v", err),
					},
				},
				IsError: true,
			},
			ID: req.ID,
		}
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      req.ID,
	}
}

// Tool implementations

type RenderArgs struct {
	File string `json:"file"`
}

func (s *Server) renderManifest(ctx context.Context, args json.RawMessage) (ToolCallResult, error) {
	var params RenderArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return ToolCallResult{}, err
	}

	// Security: Validate file path
	if err := s.security.ValidatePath(params.File); err != nil {
		s.auditor.Log("render_manifest", "read", params.File, false, err.Error())
		return ToolCallResult{}, SanitizeError(err)
	}

	// Security: Check file size
	if err := s.security.ValidateFileSize(params.File); err != nil {
		s.auditor.Log("render_manifest", "read", params.File, false, err.Error())
		return ToolCallResult{}, err
	}

	s.auditor.Log("render_manifest", "read", params.File, true, "access granted")

	docs, err := manifest.FromFiles([]string{params.File})
	if err != nil || len(docs) == 0 {
		return ToolCallResult{}, fmt.Errorf("failed to load manifest: %w", err)
	}
	doc := docs[0]

	buf := &bytes.Buffer{}
	renderer := manifest.NewRenderer(logging.NewRawLoggerTo(buf))
	if err := renderer.RenderManifest(doc); err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to render manifest: %w", err)
	}

	return ToolCallResult{
		Content: []ContentBlock{
			{
				Type: "text",
				Text: fmt.Sprintf("Rendered manifest for %s:\n\n```yaml\n%s\n```", params.File, buf.String()),
			},
		},
	}, nil
}

type DiffArgs struct {
	File string `json:"file"`
	Ref  string `json:"ref,omitempty"`
}

func (s *Server) diffManifest(ctx context.Context, args json.RawMessage) (ToolCallResult, error) {
	var params DiffArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return ToolCallResult{}, err
	}

	if params.Ref == "" {
		params.Ref = "HEAD"
	}

	// Security: Validate file path
	if err := s.security.ValidatePath(params.File); err != nil {
		s.auditor.Log("diff_manifest", "read", params.File, false, err.Error())
		return ToolCallResult{}, SanitizeError(err)
	}

	if err := s.security.ValidateFileSize(params.File); err != nil {
		s.auditor.Log("diff_manifest", "read", params.File, false, err.Error())
		return ToolCallResult{}, err
	}

	s.auditor.Log("diff_manifest", "read", params.File, true, fmt.Sprintf("comparing with %s", params.Ref))

	docs, err := manifest.FromFiles([]string{params.File})
	if err != nil || len(docs) == 0 {
		return ToolCallResult{}, fmt.Errorf("failed to load manifest: %w", err)
	}
	doc := docs[0]

	// Create differ with pretty output (easier for AI to read)
	differ := manifest.NewDiffer(params.Ref, "high", "pretty", 3)

	if err := differ.DiffManifest(doc); err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to diff manifest: %w", err)
	}

	// Note: Differ outputs directly via logger, so we return a success message
	return ToolCallResult{
		Content: []ContentBlock{
			{
				Type: "text",
				Text: fmt.Sprintf("Diff for %s (compared to %s) has been displayed", params.File, params.Ref),
			},
		},
	}, nil
}

type ValidateArgs struct {
	File string `json:"file"`
}

func (s *Server) validateManifest(ctx context.Context, args json.RawMessage) (ToolCallResult, error) {
	var params ValidateArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return ToolCallResult{}, err
	}

	docs, err := manifest.FromFiles([]string{params.File})
	if err != nil || len(docs) == 0 {
		return ToolCallResult{
			Content: []ContentBlock{
				{
					Type: "text",
					Text: fmt.Sprintf("❌ Invalid manifest %s: %v", params.File, err),
				},
			},
		}, nil
	}
	doc := docs[0]

	// Try to render to validate further
	buf := &bytes.Buffer{}
	renderer := manifest.NewRenderer(logging.NewRawLoggerTo(buf))
	if err := renderer.RenderManifest(doc); err != nil {
		return ToolCallResult{
			Content: []ContentBlock{
				{
					Type: "text",
					Text: fmt.Sprintf("❌ Manifest %s failed to render: %v", params.File, err),
				},
			},
		}, nil
	}

	return ToolCallResult{
		Content: []ContentBlock{
			{
				Type: "text",
				Text: fmt.Sprintf("✅ Manifest %s is valid", params.File),
			},
		},
	}, nil
}

type FormatArgs struct {
	File string `json:"file"`
}

func (s *Server) formatManifest(ctx context.Context, args json.RawMessage) (ToolCallResult, error) {
	var params FormatArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return ToolCallResult{}, err
	}

	docs, err := manifest.FromFiles([]string{params.File})
	if err != nil || len(docs) == 0 {
		return ToolCallResult{}, fmt.Errorf("failed to load manifest: %w", err)
	}
	doc := docs[0]

	if err := manifest.FormatManifest(doc); err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to format manifest: %w", err)
	}

	return ToolCallResult{
		Content: []ContentBlock{
			{
				Type: "text",
				Text: fmt.Sprintf("✅ Formatted %s", params.File),
			},
		},
	}, nil
}

type ListArgs struct {
	Path string `json:"path,omitempty"`
}

func (s *Server) listManifests(ctx context.Context, args json.RawMessage) (ToolCallResult, error) {
	var params ListArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return ToolCallResult{}, err
	}

	searchPath := params.Path
	if searchPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return ToolCallResult{}, err
		}
		searchPath = cwd
	}

	var manifests []string
	extensions := []string{".yaml", ".yml", ".jsonnet", ".libsonnet"}

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		for _, ext := range extensions {
			if filepath.Ext(path) == ext {
				relPath, _ := filepath.Rel(searchPath, path)
				manifests = append(manifests, relPath)
				break
			}
		}
		return nil
	})

	if err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to list manifests: %w", err)
	}

	if len(manifests) == 0 {
		return ToolCallResult{
			Content: []ContentBlock{
				{
					Type: "text",
					Text: fmt.Sprintf("No manifests found in %s", searchPath),
				},
			},
		}, nil
	}

	text := fmt.Sprintf("Found %d manifest(s) in %s:\n\n", len(manifests), searchPath)
	for _, m := range manifests {
		text += fmt.Sprintf("- %s\n", m)
	}

	return ToolCallResult{
		Content: []ContentBlock{
			{
				Type: "text",
				Text: text,
			},
		},
	}, nil
}
