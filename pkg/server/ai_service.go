package server

import (
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/vertexai/genai"
	slogcontext "github.com/PumpkinSeed/slog-context"
	api "github.com/kartverket/skipctl/pkg/api/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/api/option"
)

var (
	aiRequestsProcessed prometheus.Counter
	aiRequestsOK        prometheus.Counter
	aiRequestsFailed    prometheus.Counter
)

type AIService struct {
	api.UnimplementedAIServiceServer
	globalTimeout time.Duration
	projectID     string
	location      string
	client        *genai.Client
}

func NewAIService(ctx context.Context, reg *prometheus.Registry, globalTimeout time.Duration, projectID, location string, opts ...option.ClientOption) (*AIService, error) {
	client, err := genai.NewClient(ctx, projectID, location, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create vertex ai client: %w", err)
	}

	defineAIMetrics(reg)

	return &AIService{
		globalTimeout: globalTimeout,
		projectID:     projectID,
		location:      location,
		client:        client,
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

func (s *AIService) AnalyzeFile(stream api.AIService_AnalyzeFileServer) error {
	ctx := stream.Context()
	reqCtx := slogcontext.WithValue(ctx, "service", "ai_analyze_file")

	defer aiRequestsProcessed.Inc()

	log.InfoContext(reqCtx, "received file analysis request")

	// Create timeout context
	netCtx, cancel := globalTimeoutContext(reqCtx, s.globalTimeout)
	defer cancel()

	// Receive file chunks from client
	var fileData []byte
	var fileName string
	var mimeType string
	var prompt string

	for {
		req, err := stream.Recv()
		if err == io.EOF {
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
	response, err := s.analyzeWithVertexAI(netCtx, fileData, mimeType, prompt)
	if err != nil {
		aiRequestsFailed.Inc()
		log.ErrorContext(reqCtx, "vertex ai analysis failed", "error", err)
		return fmt.Errorf("vertex ai analysis failed: %w", err)
	}

	aiRequestsOK.Inc()
	log.InfoContext(reqCtx, "file analysis completed successfully")

	// Send response back to client
	return stream.SendAndClose(&api.AnalyzeFileResponse{
		Response: response,
	})
}

func (s *AIService) analyzeWithVertexAI(ctx context.Context, fileData []byte, mimeType, prompt string) (string, error) {
	model := s.client.GenerativeModel("gemini-1.5-flash")

	// Create file part
	filePart := genai.FileData{
		MIMEType: mimeType,
		FileURI:  "", // For inline data
	}

	// If you need to upload file first, use:
	// file, err := s.client.UploadFile(ctx, "", bytes.NewReader(fileData), &genai.UploadFileOptions{
	// 	MIMEType: mimeType,
	// })

	// Generate content
	resp, err := model.GenerateContent(ctx, genai.Text(prompt), filePart)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from vertex ai")
	}

	// Extract text response
	var result string
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			result += string(text)
		}
	}

	return result, nil
}

func (s *AIService) Close() error {
	return s.client.Close()
}
