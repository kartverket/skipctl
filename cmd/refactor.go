package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kartverket/skipctl/pkg/ai"
	"github.com/kartverket/skipctl/pkg/vectordb"
	"github.com/spf13/cobra"
)

var (
	refactorInput        string
	refactorOutput       string
	refactorModel        string
	refactorDryRun       bool
	refactorInstructions string
	refactorChromaURL    string
	refactorCollection   string
)

const refactorSystemPrompt = `You are Mai, an expert in Kubernetes manifests and ArgoKit refactoring.

Your task is to refactor Kubernetes manifests into ArgoKit format using the provided documentation and examples.

## CRITICAL: ArgoKit is a Custom DSL
ArgoKit is a domain-specific Jsonnet library created by this team. You MUST use the provided documentation 
and examples as the authoritative source for:
- Available functions and APIs
- Correct syntax and patterns
- Best practices and conventions
- Real working examples

DO NOT invent or assume ArgoKit APIs. ONLY use patterns shown in the provided documentation.

## Your refactoring approach:
1. Study the provided ArgoKit examples carefully
2. Analyze the input manifest structure
3. Identify key resources (Deployment, Service, Ingress, etc.)
4. Map to ArgoKit patterns from the documentation
5. Apply team-specific conventions from examples

## Guidelines:
- Follow the EXACT patterns from provided examples
- Use the same function names and structure as shown
- Preserve all functionality from the original manifest
- Add missing best practices shown in examples
- Explain significant changes in comments
- Keep the code readable and maintainable

Provide ONLY the Jsonnet code without additional explanation unless errors occur.`

var refactorCmd = &cobra.Command{
	Use:   "refactor [input-file]",
	Short: "Refactor Kubernetes manifests to ArgoKit format using AI",
	Long: `Refactor Kubernetes manifests to ArgoKit format using Claude AI with vector database context.

ArgoKit is a domain-specific Jsonnet library, so this command REQUIRES Chroma vector 
database with indexed ArgoKit documentation to generate accurate code.

Setup (one-time):
  1. Start Chroma:
     docker-compose -f docker-compose.chroma.yml up -d

  2. Index your ArgoKit docs:
     skipctl chroma index ./argokit-knowledge --collection argokit

Usage:
  # Refactor with ArgoKit context
  skipctl refactor deployment.yaml -o app.jsonnet

  # Preview output
  skipctl refactor deployment.yaml --dry-run

  # Use different collection
  skipctl refactor deployment.yaml --collection team-patterns -o app.jsonnet

  # Custom instructions
  skipctl refactor deployment.yaml --instructions "Add strict security policies"

Requirements:
  - ANTHROPIC_API_KEY environment variable must be set
  - Chroma vector database running (default: http://localhost:8000)
  - ArgoKit documentation indexed in Chroma collection
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		refactorInput = args[0]

		// Validate API key
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return fmt.Errorf(`No API key found. Please set:
  export ANTHROPIC_API_KEY=sk-ant-...
  
Get your API key from: https://console.anthropic.com/`)
		}

		// Validate input file exists
		if _, err := os.Stat(refactorInput); os.IsNotExist(err) {
			return fmt.Errorf("input file not found: %s", refactorInput)
		}

		// Read input manifest
		inputContent, err := os.ReadFile(refactorInput)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}

		// Create context for API calls
		ctx := context.Background()

		// Setup Chroma vector DB (required for ArgoKit context)
		chromaURL := refactorChromaURL
		if chromaURL == "" {
			chromaURL = "http://localhost:8000"
		}

		chromaClient := vectordb.NewLocalChromaClient(chromaURL)

		// Check if Chroma is available (REQUIRED)
		if !chromaClient.IsAvailable(ctx) {
			return fmt.Errorf(`❌ Chroma vector database is required for ArgoKit refactoring.

ArgoKit is a domain-specific Jsonnet library and requires documentation context.

Please start Chroma:
  docker-compose -f docker-compose.chroma.yml up -d

Then index your ArgoKit documentation:
  skipctl chroma index ./argokit-knowledge --collection %s

Or use the example docs:
  skipctl chroma index ./argokit-knowledge --collection %s

Then try refactoring again.`, refactorCollection, refactorCollection)
		}

		fmt.Println("🔍 Retrieving ArgoKit context from vector database...")

		// Query for relevant ArgoKit patterns
		query := fmt.Sprintf("ArgoKit Jsonnet Kubernetes patterns examples %s", filepath.Base(refactorInput))
		docs, err := chromaClient.Query(ctx, refactorCollection, query, 5)
		if err != nil {
			return fmt.Errorf("failed to query vector database: %w", err)
		}

		if len(docs) == 0 {
			return fmt.Errorf(`No ArgoKit documentation found in collection '%s'.

Please index your ArgoKit documentation first:
  skipctl chroma index ./argokit-knowledge --collection %s

Or use the example docs in the repo:
  skipctl chroma index ./argokit-knowledge --collection %s`, refactorCollection, refactorCollection, refactorCollection)
		}

		vectorContext := strings.Join(docs, "\n\n---\n\n")
		fmt.Printf("Retrieved %d relevant ArgoKit patterns\n", len(docs))

		// Build system prompt with ArgoKit context
		systemPrompt := refactorSystemPrompt
		systemPrompt += fmt.Sprintf("\n\n## ArgoKit Documentation and Examples:\n\n%s", vectorContext)

		if refactorInstructions != "" {
			systemPrompt += fmt.Sprintf("\n\nAdditional instructions: %s", refactorInstructions)
		}

		// Create AI client
		fmt.Println("🔄 Analyzing manifest and generating ArgoKit code...")

		aiClient := ai.NewClaudeClientWithSystemPrompt(apiKey, refactorModel, systemPrompt)

		// Build refactoring prompt
		prompt := fmt.Sprintf("Please refactor this Kubernetes manifest to ArgoKit format:\n\n```yaml\n%s\n```\n\nProvide the refactored ArgoKit Jsonnet code.", string(inputContent))

		// Call Claude
		resp, err := aiClient.SendMessage(ctx, []ai.Message{
			{
				Role: "user",
				Content: []ai.ContentItem{
					{
						Type: "text",
						Text: prompt,
					},
				},
			},
		}, nil)

		if err != nil {
			return fmt.Errorf("AI refactoring failed: %w", err)
		}

		// Extract response
		var refactoredCode string
		for _, content := range resp.Content {
			if content.Type == "text" {
				refactoredCode += content.Text
			}
		}

		if refactoredCode == "" {
			return fmt.Errorf("no refactored code received from AI")
		}

		// Clean up code markers if present
		refactoredCode = cleanCodeBlock(refactoredCode)

		// Dry-run: print to stdout
		if refactorDryRun {
			fmt.Println("\nRefactored ArgoKit code (dry-run):\n")
			fmt.Println(refactoredCode)
			return nil
		}

		// Determine output file
		outputFile := refactorOutput
		if outputFile == "" {
			// Default: same directory, change extension to .jsonnet
			base := strings.TrimSuffix(refactorInput, filepath.Ext(refactorInput))
			outputFile = base + ".argokit.jsonnet"
		}

		// Ensure output directory exists
		outputDir := filepath.Dir(outputFile)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Write output file
		if err := os.WriteFile(outputFile, []byte(refactoredCode), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}

		fmt.Printf("Refactored manifest written to: %s\n", outputFile)
		fmt.Printf("Input tokens: %d, Output tokens: %d\n", resp.Usage.InputTokens, resp.Usage.OutputTokens)

		return nil
	},
}

// cleanCodeBlock removes markdown code block markers
func cleanCodeBlock(text string) string {
	text = strings.TrimSpace(text)

	// Remove starting ```jsonnet or ```
	if strings.HasPrefix(text, "```jsonnet") {
		text = strings.TrimPrefix(text, "```jsonnet")
	} else if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```")
	}

	// Remove ending ```
	if strings.HasSuffix(text, "```") {
		text = strings.TrimSuffix(text, "```")
	}

	return strings.TrimSpace(text)
}

func init() {
	rootCmd.AddCommand(refactorCmd)

	refactorCmd.Flags().StringVarP(&refactorOutput, "output", "o", "", "Output file path (default: <input>.argokit.jsonnet)")
	refactorCmd.Flags().BoolVar(&refactorDryRun, "dry-run", false, "Preview output without writing file")
	refactorCmd.Flags().StringVar(&refactorModel, "model", "", "Claude model to use (default: claude-3-haiku-20240307)")
	refactorCmd.Flags().StringVar(&refactorInstructions, "instructions", "", "Additional refactoring instructions")
	refactorCmd.Flags().StringVar(&refactorChromaURL, "chroma-url", "http://localhost:8000", "Chroma vector database URL")
	refactorCmd.Flags().StringVar(&refactorCollection, "collection", "argokit", "Chroma collection name with ArgoKit docs")
}
