package server

import (
	"context"
	"errors"
	"fmt"
	"io"
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

func (s *AIService) RefactorToArgokitv2(stream api.AIService_RefactorToArgokitv2Server) error {
	ctx := stream.Context()
	reqCtx := slogcontext.WithValue(ctx, "service", "ai_refactor_to_argokitv2")

	defer aiRequestsProcessed.Inc()

	log.InfoContext(reqCtx, "received refactor to argokitv2 request")

	// Create timeout context
	_, cancel := globalTimeoutContext(reqCtx, s.globalTimeout)
	defer cancel()

	// Receive file chunks from client
	var fileData []byte
	var fileName string
	var mimeType string
	var prompt string

	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			aiRequestsFailed.Inc()
			log.ErrorContext(reqCtx, "failed to receive file chunk", "error", err)
			return fmt.Errorf("failed to receive file chunk: %w", err)
		}

		// First chunk contains metadata
		if fileName == "" {
			fileName = req.GetFileName()
			mimeType = req.GetMimeType()
			prompt = req.GetPrompt()
			log.InfoContext(reqCtx, "file metadata received",
				"fileName", fileName,
				"mimeType", mimeType,
				"promptLength", len(prompt))
		}

		fileData = append(fileData, req.GetChunk()...)
	}

	log.InfoContext(reqCtx, "file received", "size", len(fileData))

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
						Text: prompt,
					},
				},
			},
		},
	}

	// Send the request to Vertex AI
	resp, err := s.aiplatformService.Projects.Locations.Publishers.Models.GenerateContent(endpoint, req).Do()
	if err != nil {
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
