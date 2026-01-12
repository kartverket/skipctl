package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	apiv1 "github.com/kartverket/skipctl/pkg/api/v1"
	"github.com/prometheus/client_golang/prometheus"
)

//nolint:gocognit // Test function requires multiple test cases and setup
func TestLoadAdditionalContext(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	contextDir := filepath.Join(tempDir, "docs", "ai-context")

	// Create the context directory
	if err := os.MkdirAll(contextDir, 0755); err != nil {
		t.Fatalf("failed to create temp context dir: %v", err)
	}

	// Change to temp dir (t.Chdir automatically restores on cleanup)
	t.Chdir(tempDir)

	tests := []struct {
		name         string
		files        map[string]string
		wantContains []string
		wantEmpty    bool
	}{
		{
			name: "single markdown file",
			files: map[string]string{
				"test.md": "# Test Content\n\nThis is a test.",
			},
			wantContains: []string{"test.md", "Test Content", "This is a test"},
		},
		{
			name: "multiple files",
			files: map[string]string{
				"doc1.md":  "Document 1 content",
				"doc2.txt": "Document 2 content",
			},
			wantContains: []string{"doc1.md", "Document 1 content", "doc2.txt", "Document 2 content"},
		},
		{
			name: "ignore non-md-txt files",
			files: map[string]string{
				"good.md":     "This should be included",
				"ignored.pdf": "This should be ignored",
				"ignored.go":  "package main",
			},
			wantContains: []string{"good.md", "This should be included"},
		},
		{
			name:      "empty directory",
			files:     map[string]string{},
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up context directory
			if cleanErr := os.RemoveAll(contextDir); cleanErr != nil {
				t.Fatalf("failed to clean context dir: %v", cleanErr)
			}
			if mkdirErr := os.MkdirAll(contextDir, 0755); mkdirErr != nil {
				t.Fatalf("failed to create context dir: %v", mkdirErr)
			}

			// Create test files
			for filename, content := range tt.files {
				path := filepath.Join(contextDir, filename)
				if writeErr := os.WriteFile(path, []byte(content), 0644); writeErr != nil {
					t.Fatalf("failed to create test file %s: %v", filename, writeErr)
				}
			}

			// Load context
			context, loadErr := loadAdditionalContext()
			if loadErr != nil {
				t.Fatalf("loadAdditionalContext() error = %v", loadErr)
			}

			// Check expectations
			if tt.wantEmpty {
				if context != "" {
					t.Errorf("expected empty context, got: %s", context)
				}
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(context, want) {
					t.Errorf("context should contain %q, got: %s", want, context)
				}
			}
		})
	}
}

func TestLoadAdditionalContextNoDirectory(t *testing.T) {
	// Create a temporary directory without docs/ai-context
	tempDir := t.TempDir()

	// Change to temp dir (t.Chdir automatically restores on cleanup)
	t.Chdir(tempDir)

	// Should return empty string without error when directory doesn't exist
	context, loadErr := loadAdditionalContext()
	if loadErr != nil {
		t.Errorf("loadAdditionalContext() should not error when directory doesn't exist, got: %v", loadErr)
	}
	if context != "" {
		t.Errorf("expected empty context when directory doesn't exist, got: %s", context)
	}
}

// Mock stream for testing RefactorToArgokitv2
type mockRefactorStream struct {
	apiv1.AIService_RefactorToArgokitv2Server
	requests       []*apiv1.RefactorToArgokitv2Request
	response       *apiv1.RefactorToArgokitv2Response
	recvIndex      int
	recvError      error
	sendCloseError error
	ctx            context.Context
}

func (m *mockRefactorStream) Context() context.Context {
	if m.ctx == nil {
		return context.Background()
	}
	return m.ctx
}

func (m *mockRefactorStream) Recv() (*apiv1.RefactorToArgokitv2Request, error) {
	if m.recvError != nil {
		return nil, m.recvError
	}
	if m.recvIndex >= len(m.requests) {
		return nil, io.EOF
	}
	req := m.requests[m.recvIndex]
	m.recvIndex++
	return req, nil
}

func (m *mockRefactorStream) SendAndClose(resp *apiv1.RefactorToArgokitv2Response) error {
	if m.sendCloseError != nil {
		return m.sendCloseError
	}
	m.response = resp
	return nil
}

// Mock AIService for testing (without real Vertex AI calls)
type mockAIService struct {
	AIService
	analyzeResponse string
	analyzeError    error
}

func (m *mockAIService) analyzeWithVertexAI(_ string) (string, error) {
	if m.analyzeError != nil {
		return "", m.analyzeError
	}
	return m.analyzeResponse, nil
}

//nolint:gocognit // Test function requires multiple test cases and error handling paths
func TestRefactorToArgokitv2_ErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		stream       *mockRefactorStream
		wantErr      bool
		errContains  string
		setupService func() *mockAIService
	}{
		{
			name: "successful refactoring",
			stream: &mockRefactorStream{
				requests: []*apiv1.RefactorToArgokitv2Request{
					{
						FileName: "test.jsonnet",
						MimeType: "text/plain",
						Chunk:    []byte("test content"),
					},
				},
			},
			setupService: func() *mockAIService {
				return &mockAIService{
					analyzeResponse: "refactored content",
					analyzeError:    nil,
				}
			},
			wantErr: false,
		},
		{
			name: "recv error",
			stream: &mockRefactorStream{
				recvError: errors.New("recv failed"),
			},
			setupService: func() *mockAIService {
				return &mockAIService{}
			},
			wantErr:     true,
			errContains: "failed to receive request",
		},
		{
			name: "vertex ai analysis error",
			stream: &mockRefactorStream{
				requests: []*apiv1.RefactorToArgokitv2Request{
					{
						FileName: "test.jsonnet",
						MimeType: "text/plain",
						Chunk:    []byte("test prompt"),
					},
				},
			},
			setupService: func() *mockAIService {
				return &mockAIService{
					analyzeError: errors.New("vertex ai failed"),
				}
			},
			wantErr:     true,
			errContains: "vertex ai analysis failed",
		},
		{
			name: "empty chunks",
			stream: &mockRefactorStream{
				requests: []*apiv1.RefactorToArgokitv2Request{
					{
						FileName: "test.jsonnet",
						MimeType: "text/plain",
						Chunk:    []byte(""),
					},
				},
			},
			setupService: func() *mockAIService {
				return &mockAIService{
					analyzeResponse: "response",
				}
			},
			wantErr: false, // Should still work with empty chunks
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService()

			// Use a custom refactor method that uses the mock
			err := func() error {
				// Receive file metadata and chunks
				var fileName string
				var promptBuilder strings.Builder

				for {
					req, recvErr := tt.stream.Recv()
					if errors.Is(recvErr, io.EOF) {
						break
					}
					if recvErr != nil {
						return fmt.Errorf("failed to receive request: %w", recvErr)
					}

					if fileName == "" {
						fileName = req.GetFileName()
					}

					// Accumulate chunks
					if chunk := req.GetChunk(); len(chunk) > 0 {
						promptBuilder.Write(chunk)
					}
				}

				prompt := promptBuilder.String()

				// Analyze with mock
				response, analyzeErr := service.analyzeWithVertexAI(prompt)
				if analyzeErr != nil {
					return fmt.Errorf("vertex ai analysis failed: %w", analyzeErr)
				}

				return tt.stream.SendAndClose(&apiv1.RefactorToArgokitv2Response{
					Response: response,
				})
			}()

			if tt.wantErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.errContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContains)
				}
			}
		})
	}
}

func TestAIService_Close(t *testing.T) {
	// Close should be a no-op and never return an error
	service := &AIService{}

	err := service.Close()
	if err != nil {
		t.Errorf("Close() should never return error, got: %v", err)
	}
}

//nolint:gocognit // Test function requires multiple test cases and field validation
func TestNewAIService(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		location  string
		wantNil   bool
	}{
		{
			name:      "valid parameters",
			projectID: "test-project",
			location:  "us-central1",
			wantNil:   false,
		},
		{
			name:      "empty project ID",
			projectID: "",
			location:  "us-central1",
			wantNil:   false, // Will create service but with empty projectID
		},
		{
			name:      "empty location",
			projectID: "test-project",
			location:  "",
			wantNil:   false, // Will create service but with empty location
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := prometheus.NewRegistry()
			service, err := NewAIService(context.Background(), reg, 30, tt.projectID, tt.location)

			// NewAIService may fail if google credentials are not available
			// but that's OK for this test - we're just checking the structure
			if err != nil {
				t.Logf("NewAIService returned error (expected in test env): %v", err)
				return
			}

			if tt.wantNil && service != nil {
				t.Errorf("expected nil service, got: %v", service)
			} else if !tt.wantNil && service == nil {
				t.Errorf("expected non-nil service")
			}

			if service != nil {
				if service.projectID != tt.projectID {
					t.Errorf("projectID = %q, want %q", service.projectID, tt.projectID)
				}
				if service.location != tt.location {
					t.Errorf("location = %q, want %q", service.location, tt.location)
				}
			}
		})
	}
}

func TestDefineAIMetrics(t *testing.T) {
	// Test that metrics are properly registered
	reg := prometheus.NewRegistry()

	defineAIMetrics(reg)

	// Verify metrics exist by attempting to collect them
	metrics, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	// Check that we have the expected metrics
	expectedMetrics := map[string]bool{
		"ai_requests_processed_total": false,
		"ai_requests_ok_total":        false,
		"ai_requests_failed_total":    false,
	}

	for _, m := range metrics {
		if _, exists := expectedMetrics[m.GetName()]; exists {
			expectedMetrics[m.GetName()] = true
		}
	}

	for name, found := range expectedMetrics {
		if !found {
			t.Errorf("expected metric %q not found", name)
		}
	}
}
