package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	slogcontext "github.com/PumpkinSeed/slog-context"
	api "github.com/kartverket/skipctl/pkg/api/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/aiplatform/v1"
	"google.golang.org/api/option"
)

var (
	aiRequestsProcessed prometheus.Counter
	aiRequestsOK        prometheus.Counter
	aiRequestsFailed    prometheus.Counter
)

type AIService struct {
	api.UnimplementedAIServiceServer
	globalTimeout     time.Duration
	projectID         string
	location          string
	model             string
	aiplatformService *aiplatform.Service
}

func NewAIService(ctx context.Context, reg *prometheus.Registry, globalTimeout time.Duration, projectID, location string) (*AIService, error) {
	// Create authenticated client using Application Default Credentials
	client, err := google.DefaultClient(ctx, aiplatform.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("failed to create google default client: %w", err)
	}

	aiplatformService, err := aiplatform.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create aiplatform service: %w", err)
	}

	defineAIMetrics(reg)

	return &AIService{
		globalTimeout:     globalTimeout,
		projectID:         projectID,
		location:          location,
		model:             "gemini-2.5-flash-lite",
		aiplatformService: aiplatformService,
	}, nil
}

func defineAIMetrics(reg *prometheus.Registry) {
	aiRequestsProcessed = promauto.With(reg).NewCounter(prometheus.CounterOpts{
		Name: "ai_requests_processed_total",
		Help: "The total number of processed AI requests",
	})
	aiRequestsOK = promauto.With(reg).NewCounter(prometheus.CounterOpts{
		Name: "ai_requests_ok_total",
		Help: "The total number of successful AI requests",
	})
	aiRequestsFailed = promauto.With(reg).NewCounter(prometheus.CounterOpts{
		Name: "ai_requests_failed_total",
		Help: "The total number of failed AI requests",
	})
}

// loadAdditionalContext reads all .md and .txt files from docs/ai-context/
// and concatenates their contents to be added to the AI prompt
func loadAdditionalContext() (string, error) {
	contextDir := "docs/ai-context"

	// Check if directory exists
	if _, err := os.Stat(contextDir); os.IsNotExist(err) {
		// Directory doesn't exist, return empty string (not an error)
		if log != nil {
			log.Info("ai-context directory does not exist, skipping additional context")
		}
		return "", nil
	}

	var contextBuilder strings.Builder
	var filesLoaded []string

	// Walk through the directory
	err := filepath.Walk(contextDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-.md/.txt files
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".txt" {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		filesLoaded = append(filesLoaded, filepath.Base(path))

		// Add to context with a separator
		contextBuilder.WriteString(fmt.Sprintf("\n\n--- Additional Context from %s ---\n\n", filepath.Base(path)))
		contextBuilder.Write(content)

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to load additional context: %w", err)
	}

	if len(filesLoaded) > 0 && log != nil {
		log.Info("loaded additional context files", "files", filesLoaded, "totalSize", contextBuilder.Len())
	}

	return contextBuilder.String(), nil
}

func (s *AIService) RefactorToArgokitv2(stream api.AIService_RefactorToArgokitv2Server) error {
	ctx := stream.Context()
	reqCtx := slogcontext.WithValue(ctx, "service", "ai_refactor_to_argokitv2")

	defer aiRequestsProcessed.Inc()

	log.InfoContext(reqCtx, "received refactor to argokitv2 request")

	// Receive file metadata and chunks from client
	var fileName string
	var mimeType string
	var promptBuilder strings.Builder

	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			aiRequestsFailed.Inc()
			log.ErrorContext(reqCtx, "failed to receive request", "error", err)
			return fmt.Errorf("failed to receive request: %w", err)
		}

		// First chunk contains metadata
		if fileName == "" {
			fileName = req.GetFileName()
			mimeType = req.GetMimeType()
			log.InfoContext(reqCtx, "file metadata received",
				"fileName", fileName,
				"mimeType", mimeType)
		}

		// Accumulate all chunks to build the prompt
		if chunk := req.GetChunk(); len(chunk) > 0 {
			promptBuilder.Write(chunk)
		}
	}

	prompt := promptBuilder.String()
	log.InfoContext(reqCtx, "prompt assembled from chunks", "promptLength", len(prompt))

	// Analyze file with Vertex AI
	response, err := s.analyzeWithVertexAI(prompt)
	if err != nil {
		aiRequestsFailed.Inc()
		log.ErrorContext(reqCtx, "vertex ai analysis failed", "error", err)
		return fmt.Errorf("vertex ai analysis failed: %w", err)
	}

	aiRequestsOK.Inc()
	log.InfoContext(reqCtx, "refactor to argokitv2 completed successfully")

	// Send response back to client
	return stream.SendAndClose(&api.RefactorToArgokitv2Response{
		Response: response,
	})
}

func (s *AIService) analyzeWithVertexAI(prompt string) (string, error) {
	// Create a timeout context for the Vertex AI API call
	ctx, cancel := context.WithTimeout(context.Background(), s.globalTimeout)
	defer cancel()

	// Load additional context from docs/ai-context/
	additionalContext, err := loadAdditionalContext()
	if err != nil {
		// Log warning but continue without additional context
		if log != nil {
			log.Warn("failed to load additional context", "error", err)
		}
		additionalContext = ""
	}

	// Combine the prompt with additional context
	finalPrompt := prompt
	if additionalContext != "" {
		finalPrompt = prompt + additionalContext
		if log != nil {
			log.Info("added additional context to prompt", "additionalContextLength", len(additionalContext))
		}
	}

	// Construct the endpoint for the model
	endpoint := fmt.Sprintf("projects/%s/locations/%s/publishers/google/models/%s",
		s.projectID, s.location, s.model)

	// Create the request
	req := &aiplatform.GoogleCloudAiplatformV1GenerateContentRequest{
		Contents: []*aiplatform.GoogleCloudAiplatformV1Content{
			{
				Role: "user",
				Parts: []*aiplatform.GoogleCloudAiplatformV1Part{
					{
						Text: finalPrompt,
					},
				},
			},
		},
	}

	// Send the request to Vertex AI with timeout context
	resp, err := s.aiplatformService.Projects.Locations.Publishers.Models.GenerateContent(endpoint, req).Context(ctx).Do()
	if err != nil {
		// Check if it was a timeout error
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("vertex ai request timed out after %v: %w", s.globalTimeout, err)
		}
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	// Extract the response text
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("no response from vertex ai")
	}

	if text := resp.Candidates[0].Content.Parts[0].Text; text != "" {
		return text, nil
	}

	return "", errors.New("empty response from vertex ai")
}

func (s *AIService) Close() error {
	// No explicit cleanup is required here. The aiplatform.Service uses an HTTP
	// client provided by google.DefaultClient, whose resources (connections,
	// goroutines, etc.) are managed by the Go runtime and do not need an
	// explicit Close call. This method exists to satisfy the AIService interface
	// and is intentionally a no-op.
	return nil
}
