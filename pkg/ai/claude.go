package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	claudeAPIURL     = "https://api.anthropic.com/v1/messages"
	claudeAPIVersion = "2023-06-01"
)

// ClaudeClient handles communication with Claude API
type ClaudeClient struct {
	apiKey       string
	model        string
	httpClient   *http.Client
	systemPrompt string
}

// NewClaudeClient creates a new Claude API client
func NewClaudeClient(apiKey string, model string) *ClaudeClient {
	if model == "" {
		// Use Claude 3 Haiku (fast and cost-effective)
		model = "claude-3-haiku-20240307"
	}
	return &ClaudeClient{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// NewClaudeClientWithSystemPrompt creates a client with custom system prompt
func NewClaudeClientWithSystemPrompt(apiKey string, model string, systemPrompt string) *ClaudeClient {
	if model == "" {
		model = "claude-3-haiku-20240307"
	}
	return &ClaudeClient{
		apiKey:       apiKey,
		model:        model,
		httpClient:   &http.Client{Timeout: 60 * time.Second},
		systemPrompt: systemPrompt,
	}
}

// SetSystemPrompt updates the system prompt
func (c *ClaudeClient) SetSystemPrompt(prompt string) {
	c.systemPrompt = prompt
}

// Message represents a message in the conversation
type Message struct {
	Role    string        `json:"role"`
	Content []ContentItem `json:"content"`
}

// ContentItem represents different types of content
type ContentItem struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	ID        string                 `json:"id,omitempty"`          // For tool_use requests
	ToolUseID string                 `json:"tool_use_id,omitempty"` // For tool_result responses
	Name      string                 `json:"name,omitempty"`
	Input     map[string]interface{} `json:"input,omitempty"`
	Content   string                 `json:"content,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for ContentItem
func (c ContentItem) MarshalJSON() ([]byte, error) {
	type Alias ContentItem
	// For tool_use, always include input field even if empty
	if c.Type == "tool_use" {
		if c.Input == nil {
			c.Input = make(map[string]interface{})
		}
		// Create a map to ensure input is always present
		return json.Marshal(&struct {
			Type  string                 `json:"type"`
			ID    string                 `json:"id,omitempty"`
			Name  string                 `json:"name,omitempty"`
			Input map[string]interface{} `json:"input"`
		}{
			Type:  c.Type,
			ID:    c.ID,
			Name:  c.Name,
			Input: c.Input,
		})
	}
	return json.Marshal((Alias)(c))
}

// Tool represents a tool that Claude can use
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// Request represents a request to Claude API
type Request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
	Tools     []Tool    `json:"tools,omitempty"`
	System    string    `json:"system,omitempty"`
}

// Response represents Claude API response
type Response struct {
	ID      string        `json:"id"`
	Type    string        `json:"type"`
	Role    string        `json:"role"`
	Content []ContentItem `json:"content"`
	Model   string        `json:"model"`
	Usage   struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	StopReason string `json:"stop_reason"`
}

// SendMessage sends a message to Claude and returns the response
func (c *ClaudeClient) SendMessage(ctx context.Context, messages []Message, tools []Tool) (*Response, error) {
	req := Request{
		Model:     c.model,
		MaxTokens: 4096,
		Messages:  messages,
		Tools:     tools,
		System:    c.systemPrompt,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", claudeAPIURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", claudeAPIVersion)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Fix tool_use items that don't have input field
	for i, content := range response.Content {
		if content.Type == "tool_use" && content.Input == nil {
			response.Content[i].Input = make(map[string]interface{})
		}
	}

	return &response, nil
}
