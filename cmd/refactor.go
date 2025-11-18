package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kartverket/skipctl/pkg/ai"
	"github.com/spf13/cobra"
)

var (
	refactorInput        string
	refactorOutput       string
	refactorModel        string
	refactorVectorDB     string
	refactorDryRun       bool
	refactorInstructions string
)

const refactorSystemPrompt = `You are Mai, an expert in Kubernetes manifests and ArgoKit refactoring.

Your task is to refactor Kubernetes manifests into ArgoKit format.

## ArgoKit Overview:
ArgoKit is a structured way to define Kubernetes applications using Jsonnet. It provides:
- Consistent application structure
- Reusable components
- Type-safe configuration
- Built-in best practices

## Your refactoring approach:
1. Analyze the input manifest structure
2. Identify key resources (Deployment, Service, Ingress, etc.)
3. Convert to ArgoKit Application format using Jsonnet
4. Apply ArgoKit best practices:
   - Use proper resource limits
   - Configure health checks
   - Set up monitoring labels
   - Apply security contexts
   - Use namespaces correctly

## Output format:
Provide the refactored ArgoKit manifest as valid Jsonnet code. Include:
- Application metadata
- Container configuration
- Service definitions
- Ingress rules (if applicable)
- Resource limits
- Environment variables
- ConfigMaps/Secrets references

## Guidelines:
- Preserve all functionality from the original manifest
- Add missing best practices (resource limits, health checks)
- Use ArgoKit idioms and patterns
- Explain significant changes in comments
- Keep the code readable and maintainable

Provide ONLY the Jsonnet code without additional explanation unless errors occur.`

var refactorCmd = &cobra.Command{
	Use:   "refactor [input-file]",
	Short: "Refactor Kubernetes manifests to ArgoKit format using AI",
	Long: `Refactor Kubernetes manifests to ArgoKit format using Claude AI.

This command takes a Kubernetes manifest (YAML/JSON) and converts it to ArgoKit's
Jsonnet-based format with best practices applied.

Examples:
  # Refactor a manifest and write to file
  skipctl refactor deployment.yaml -o argokit/app.jsonnet

  # Preview refactored output (dry-run)
  skipctl refactor deployment.yaml --dry-run

  # Use custom instructions
  skipctl refactor deployment.yaml --instructions "Focus on security best practices"

  # Use vector DB for enhanced context (when available)
  skipctl refactor deployment.yaml --vector-db ./docs

Requirements:
  - ANTHROPIC_API_KEY environment variable must be set
  - Input file must be a valid Kubernetes manifest
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

		// Prepare system prompt
		systemPrompt := refactorSystemPrompt
		if refactorInstructions != "" {
			systemPrompt += fmt.Sprintf("\n\nAdditional instructions: %s", refactorInstructions)
		}

		// Load vector DB context if provided
		var vectorContext string
		if refactorVectorDB != "" {
			vectorContext, err = loadVectorContext(refactorVectorDB)
			if err != nil {
				fmt.Fprintf(os.Stderr, "⚠️  Warning: Could not load vector DB context: %v\n", err)
			} else if vectorContext != "" {
				systemPrompt += fmt.Sprintf("\n\nRelevant ArgoKit documentation and examples:\n%s", vectorContext)
			}
		}

		// Create AI client
		fmt.Println("🔄 Analyzing manifest and generating ArgoKit code...")

		client := ai.NewClaudeClientWithSystemPrompt(apiKey, refactorModel, systemPrompt)

		// Build refactoring prompt
		prompt := fmt.Sprintf("Please refactor this Kubernetes manifest to ArgoKit format:\n\n```yaml\n%s\n```\n\nProvide the refactored ArgoKit Jsonnet code.", string(inputContent))

		// Call Claude
		ctx := context.Background()
		resp, err := client.SendMessage(ctx, []ai.Message{
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
			fmt.Println("\n✅ Refactored ArgoKit code (dry-run):\n")
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

		fmt.Printf("✅ Refactored manifest written to: %s\n", outputFile)
		fmt.Printf("📊 Input tokens: %d, Output tokens: %d\n", resp.Usage.InputTokens, resp.Usage.OutputTokens)

		return nil
	},
}

// loadVectorContext loads relevant ArgoKit documentation from vector DB
// This is a placeholder for future vector DB integration
func loadVectorContext(vectorDBPath string) (string, error) {
	// TODO: Implement vector DB query
	// For now, try to load example files from the directory

	if _, err := os.Stat(vectorDBPath); os.IsNotExist(err) {
		return "", fmt.Errorf("vector DB path does not exist: %s", vectorDBPath)
	}

	// Look for .md and .jsonnet files as examples
	var examples []string
	err := filepath.Walk(vectorDBPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext == ".md" || ext == ".jsonnet" {
			content, err := os.ReadFile(path)
			if err == nil {
				examples = append(examples, fmt.Sprintf("File: %s\n%s\n", filepath.Base(path), string(content)))
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if len(examples) == 0 {
		return "", nil
	}

	// Limit context to avoid token limits
	combined := strings.Join(examples, "\n---\n")
	if len(combined) > 10000 {
		combined = combined[:10000] + "\n... (truncated)"
	}

	return combined, nil
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
	refactorCmd.Flags().StringVar(&refactorVectorDB, "vector-db", "", "Path to vector DB with ArgoKit docs/examples")
	refactorCmd.Flags().StringVar(&refactorInstructions, "instructions", "", "Additional refactoring instructions")
}
