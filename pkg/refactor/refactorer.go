package refactor

import (
	"context"
	"fmt"
	"os"

	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/kartverket/skipctl/pkg/prompts"
	"google.golang.org/api/aiplatform/v1"
	"google.golang.org/api/option"
)

func refactorJsonnet(doc *manifest.Document) error {
	projectID := os.Getenv("GCP_PROJECT_ID") // TODO: Get from config
	if projectID == "" {
		return fmt.Errorf("there are no project id")
	}
	location := "europe-north1"   // TODO: Get from config
	model := "gemini-1.5-pro-001" // TODO: Get from config
	apiKey := os.Getenv("GCP_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("there are no GCP API key here")

	}
	ctx := context.Background()

	aiplatformService, err := aiplatform.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return fmt.Errorf("failed to create new aiplatform service: %w", err)
	}

	// Construct the request
	endpoint := fmt.Sprintf("projects/%s/locations/%s/publishers/google/models/%s", projectID, location, model)
	req := &aiplatform.GoogleCloudAiplatformV1GenerateContentRequest{
		SystemInstruction: &aiplatform.GoogleCloudAiplatformV1Content{
			Parts: []*aiplatform.GoogleCloudAiplatformV1Part{
				{
					Text: prompts.RefactorSystemPrompt,
				},
			},
		},
		Contents: []*aiplatform.GoogleCloudAiplatformV1Content{
			{
				Role: "user",
				Parts: []*aiplatform.GoogleCloudAiplatformV1Part{
					{
						Text: string(doc.Content),
					},
				},
			},
		},
	}

	// Send the request
	resp, err := aiplatformService.Projects.Locations.Publishers.Models.GenerateContent(endpoint, req).Do()
	if err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if text := resp.Candidates[0].Content.Parts[0].Text; text != "" {
			// Create a new file path for the refactored content
			ext := ".jsonnet"
			newPath := fmt.Sprintf("%s.refactored%s", "test", ext)

			// Write the refactored content to the new file
			err := os.WriteFile(newPath, []byte(text), 0644)
			if err != nil {
				return fmt.Errorf("failed to write refactored content to file: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("no content generated or unexpected response format")
}
func RefactorManifest(file *manifest.Document) error {
	return refactorJsonnet(file)
}
