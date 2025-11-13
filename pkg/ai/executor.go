package ai

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/manifest"
)

// ToolExecutor executes tools that Claude requests
type ToolExecutor struct {
	security *SecurityConfig
}

// NewToolExecutor creates a new tool executor with default security config
func NewToolExecutor() *ToolExecutor {
	return &ToolExecutor{
		security: DefaultSecurityConfig(),
	}
}

// NewToolExecutorWithSecurity creates a new tool executor with custom security config
func NewToolExecutorWithSecurity(security *SecurityConfig) *ToolExecutor {
	return &ToolExecutor{
		security: security,
	}
}

// ExecuteTool executes a tool and returns the result
func (te *ToolExecutor) ExecuteTool(ctx context.Context, toolName string, input map[string]interface{}) (string, error) {
	switch toolName {
	case "render_manifest":
		return te.renderManifest(input)
	case "diff_manifest":
		return te.diffManifest(input)
	case "validate_manifest":
		return te.validateManifest(input)
	case "format_manifest":
		return te.formatManifest(input)
	case "list_manifests":
		return te.listManifests(input)
	default:
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
}

func (te *ToolExecutor) renderManifest(input map[string]interface{}) (string, error) {
	file, ok := input["file"].(string)
	if !ok {
		return "", fmt.Errorf("file parameter required")
	}

	// Validate file access
	if err := te.security.ValidatePath(file); err != nil {
		return "", fmt.Errorf("security validation failed: %w", err)
	}

	docs, err := manifest.FromFiles([]string{file})
	if err != nil || len(docs) == 0 {
		return "", fmt.Errorf("failed to load manifest: %w", err)
	}
	doc := docs[0]

	buf := &bytes.Buffer{}
	renderer := manifest.NewRenderer(logging.NewRawLoggerTo(buf))
	if err := renderer.RenderManifest(doc); err != nil {
		return "", fmt.Errorf("failed to render manifest: %w", err)
	}

	return fmt.Sprintf("Rendered manifest for %s:\n\n```yaml\n%s\n```", file, buf.String()), nil
}

func (te *ToolExecutor) diffManifest(input map[string]interface{}) (string, error) {
	file, ok := input["file"].(string)
	if !ok {
		return "", fmt.Errorf("file parameter required")
	}

	// Validate file access
	if err := te.security.ValidatePath(file); err != nil {
		return "", fmt.Errorf("security validation failed: %w", err)
	}

	ref, ok := input["ref"].(string)
	if !ok || ref == "" {
		ref = "HEAD"
	}

	docs, err := manifest.FromFiles([]string{file})
	if err != nil || len(docs) == 0 {
		return "", fmt.Errorf("failed to load manifest: %w", err)
	}
	doc := docs[0]

	// Create differ with pretty output
	buf := &bytes.Buffer{}
	differ := manifest.NewDiffer(ref, "high", "pretty", 3)

	// Temporarily capture output
	oldLogger := logging.RawLogger()
	defer func() {
		// Restore original logger
		_ = oldLogger
	}()

	if err := differ.DiffManifest(doc); err != nil {
		return "", fmt.Errorf("failed to diff manifest: %w", err)
	}

	output := buf.String()
	if output == "" {
		return fmt.Sprintf("No changes detected in %s compared to %s", file, ref), nil
	}

	return fmt.Sprintf("Diff for %s (compared to %s):\n\n%s", file, ref, output), nil
}

func (te *ToolExecutor) validateManifest(input map[string]interface{}) (string, error) {
	file, ok := input["file"].(string)
	if !ok {
		return "", fmt.Errorf("file parameter required")
	}

	// Validate file access
	if err := te.security.ValidatePath(file); err != nil {
		return "", fmt.Errorf("security validation failed: %w", err)
	}

	docs, err := manifest.FromFiles([]string{file})
	if err != nil || len(docs) == 0 {
		return fmt.Sprintf("❌ Invalid manifest %s: %v", file, err), nil
	}
	doc := docs[0]

	// Try to render to validate further
	buf := &bytes.Buffer{}
	renderer := manifest.NewRenderer(logging.NewRawLoggerTo(buf))
	if err := renderer.RenderManifest(doc); err != nil {
		return fmt.Sprintf("❌ Manifest %s failed to render: %v", file, err), nil
	}

	return fmt.Sprintf("✅ Manifest %s is valid", file), nil
}

func (te *ToolExecutor) formatManifest(input map[string]interface{}) (string, error) {
	file, ok := input["file"].(string)
	if !ok {
		return "", fmt.Errorf("file parameter required")
	}

	// Validate file access
	if err := te.security.ValidatePath(file); err != nil {
		return "", fmt.Errorf("security validation failed: %w", err)
	}

	// Validate write operation
	if err := te.security.ValidateWriteOperation("format"); err != nil {
		return "", fmt.Errorf("write operation validation failed: %w", err)
	}

	docs, err := manifest.FromFiles([]string{file})
	if err != nil || len(docs) == 0 {
		return "", fmt.Errorf("failed to load manifest: %w", err)
	}
	doc := docs[0]

	if err := manifest.FormatManifest(doc); err != nil {
		return "", fmt.Errorf("failed to format manifest: %w", err)
	}

	return fmt.Sprintf("✅ Formatted %s", file), nil
}

func (te *ToolExecutor) listManifests(input map[string]interface{}) (string, error) {
	searchPath, ok := input["path"].(string)
	if !ok || searchPath == "" {
		var err error
		searchPath, err = getCurrentDir()
		if err != nil {
			return "", err
		}
	}

	// Validate path access
	if err := te.security.ValidatePath(searchPath); err != nil {
		return "", fmt.Errorf("security validation failed: %w", err)
	}

	manifests, err := findManifestFiles(searchPath)
	if err != nil {
		return "", fmt.Errorf("failed to list manifests: %w", err)
	}

	if len(manifests) == 0 {
		return fmt.Sprintf("No manifests found in %s", searchPath), nil
	}

	result := fmt.Sprintf("Found %d manifest(s) in %s:\n\n", len(manifests), searchPath)
	for _, m := range manifests {
		result += fmt.Sprintf("- %s\n", m)
	}

	return result, nil
}

// Helper functions
func getCurrentDir() (string, error) {
	return os.Getwd()
}

func findManifestFiles(searchPath string) ([]string, error) {
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

	return manifests, err
}

// Agent handles the conversation loop with Claude
type Agent struct {
	client      *ClaudeClient
	executor    *ToolExecutor
	rateLimiter *RateLimiter
	messages    []Message
}

// NewAgent creates a new AI agent with default settings
func NewAgent(apiKey string, model string) *Agent {
	return &Agent{
		client:      NewClaudeClient(apiKey, model),
		executor:    NewToolExecutor(),
		rateLimiter: NewRateLimiter(20, time.Minute), // 20 requests per minute
		messages:    []Message{},
	}
}

// NewAgentWithSecurity creates a new AI agent with custom security settings
func NewAgentWithSecurity(apiKey string, model string, security *SecurityConfig, rateLimiter *RateLimiter) *Agent {
	return &Agent{
		client:      NewClaudeClient(apiKey, model),
		executor:    NewToolExecutorWithSecurity(security),
		rateLimiter: rateLimiter,
		messages:    []Message{},
	}
}

// Ask sends a question to Claude and handles tool calls
func (a *Agent) Ask(ctx context.Context, question string) (string, error) {
	// Check rate limit
	if a.rateLimiter != nil {
		if err := a.rateLimiter.Allow(); err != nil {
			return "", fmt.Errorf("rate limit exceeded: %w", err)
		}
	}

	// Add user message
	a.messages = append(a.messages, Message{
		Role: "user",
		Content: []ContentItem{
			{
				Type: "text",
				Text: question,
			},
		},
	})

	tools := GetAvailableTools()
	maxIterations := 10 // Prevent infinite loops

	for i := 0; i < maxIterations; i++ {
		// Send to Claude
		resp, err := a.client.SendMessage(ctx, a.messages, tools)
		if err != nil {
			return "", err
		}

		// Add assistant's response to history - but fix tool_use input fields
		assistantContent := make([]ContentItem, len(resp.Content))
		copy(assistantContent, resp.Content)

		// Ensure all tool_use items have input field for next API call
		for i, content := range assistantContent {
			if content.Type == "tool_use" && content.Input == nil {
				assistantContent[i].Input = make(map[string]interface{})
			}
		}

		a.messages = append(a.messages, Message{
			Role:    "assistant",
			Content: assistantContent,
		})

		// Check if Claude wants to use tools
		var toolResults []ContentItem
		var hasToolUse bool

		for _, content := range resp.Content {
			if content.Type == "tool_use" {
				hasToolUse = true

				// Ensure input is not nil
				if content.Input == nil {
					content.Input = make(map[string]interface{})
				}

				// Execute the tool
				result, err := a.executor.ExecuteTool(ctx, content.Name, content.Input)
				if err != nil {
					result = fmt.Sprintf("Error executing tool: %v", err)
				}

				// Add tool result - use ID from tool_use as tool_use_id in response
				toolResults = append(toolResults, ContentItem{
					Type:      "tool_result",
					ToolUseID: content.ID, // Use ID from the tool_use request
					Content:   result,
				})
			}
		}

		// If Claude used tools, send results back
		if hasToolUse {
			a.messages = append(a.messages, Message{
				Role:    "user",
				Content: toolResults,
			})
			continue // Loop to get Claude's final response
		}

		// No more tools, extract final answer
		var answer string
		for _, content := range resp.Content {
			if content.Type == "text" {
				answer += content.Text
			}
		}

		return answer, nil
	}

	return "", fmt.Errorf("max iterations reached")
}

// Reset clears the conversation history
func (a *Agent) Reset() {
	a.messages = []Message{}
}
