package refactor

import (
	"context"
	"fmt"
	"os"

	"github.com/kartverket/skipctl/pkg/manifest"
	"github.com/kartverket/skipctl/pkg/prompts"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/aiplatform/v1"
	"google.golang.org/api/option"
)

func refactorJsonnet(doc *manifest.Document) error {
	ctx := context.Background()
	projectID := "kv-spire-devex-ksde" // TODO: Get from config
	if projectID == "" {
		return fmt.Errorf("there are no project id")
	}
	location := "europe-north1"      // TODO: Get from config
	model := "gemini-2.5-flash-lite" // TODO: Get from config
	client, err := google.DefaultClient(ctx, aiplatform.CloudPlatformScope)
	if err != nil {
		return fmt.Errorf("failed to create google default client: %w", err)
	}

	aiplatformService, err := aiplatform.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("failed to create new aiplatform service: %w", err)
	}
	// Construct the request
	endpoint := fmt.Sprintf("projects/%s/locations/%s/publishers/google/models/%s", projectID, location, model)
	// Replace with your actual Vertex AI Search Datastore ID and location
	vertexAISearchDatastoreID := "argokit-v2-knowledge_1764338186592"
	vertexAISearchDatastoreLocation := "eu" // Or the specific location of your datastore

	// Construct the full resource name for the Vertex AI Search datastore
	// Format: projects/{project}/locations/{location}/collections/default_collection/dataStores/{datastore_id}
	datastoreResourceName := fmt.Sprintf("projects/%s/locations/%s/collections/default_collection/dataStores/%s",
		projectID, vertexAISearchDatastoreLocation, vertexAISearchDatastoreID)

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
		Tools: []*aiplatform.GoogleCloudAiplatformV1Tool{
			{
				Retrieval: &aiplatform.GoogleCloudAiplatformV1Retrieval{
					VertexAiSearch: &aiplatform.GoogleCloudAiplatformV1VertexAISearch{
						Datastore: datastoreResourceName,
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
