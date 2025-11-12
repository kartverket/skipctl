package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestMCPServer_Initialize(t *testing.T) {
	// Add newline to match scanner expectation
	input := `{"jsonrpc":"2.0","method":"initialize","id":1}` + "\n"
	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	server := NewServer(reader, writer)
	ctx := context.Background()

	// Run server (it will exit after processing the single request)
	if err := server.Serve(ctx); err != nil {
		t.Fatalf("Server failed: %v", err)
	}

	// Decode response
	var response JSONRPCResponse
	decoder := json.NewDecoder(writer)
	if err := decoder.Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.JSONRPC != "2.0" {
		t.Errorf("Expected jsonrpc 2.0, got %s", response.JSONRPC)
	}

	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}

	// Check result structure
	result, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be a map, got %T", response.Result)
	}

	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("Expected protocol version 2024-11-05, got %v", result["protocolVersion"])
	}
}

func TestMCPServer_ToolsList(t *testing.T) {
	input := `{"jsonrpc":"2.0","method":"tools/list","id":2}` + "\n"
	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	server := NewServer(reader, writer)
	ctx := context.Background()

	if err := server.Serve(ctx); err != nil {
		t.Fatalf("Server failed: %v", err)
	}

	var response JSONRPCResponse
	decoder := json.NewDecoder(writer)
	if err := decoder.Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}

	// Check that we have tools
	result, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be a map, got %T", response.Result)
	}

	tools, ok := result["tools"].([]interface{})
	if !ok {
		t.Fatalf("Expected tools to be an array, got %T", result["tools"])
	}

	expectedTools := []string{"render_manifest", "diff_manifest", "validate_manifest", "format_manifest", "list_manifests"}
	if len(tools) != len(expectedTools) {
		t.Errorf("Expected %d tools, got %d", len(expectedTools), len(tools))
	}

	// Verify tool names
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolMap := tool.(map[string]interface{})
		name := toolMap["name"].(string)
		toolNames[name] = true
	}

	for _, expected := range expectedTools {
		if !toolNames[expected] {
			t.Errorf("Expected tool %s not found", expected)
		}
	}
}

func TestMCPServer_InvalidMethod(t *testing.T) {
	input := `{"jsonrpc":"2.0","method":"invalid_method","id":3}` + "\n"
	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	server := NewServer(reader, writer)
	ctx := context.Background()

	if err := server.Serve(ctx); err != nil {
		t.Fatalf("Server failed: %v", err)
	}

	var response JSONRPCResponse
	decoder := json.NewDecoder(writer)
	if err := decoder.Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Error == nil {
		t.Error("Expected error for invalid method")
	}

	if response.Error.Code != -32601 {
		t.Errorf("Expected error code -32601, got %d", response.Error.Code)
	}
}
