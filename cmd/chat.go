package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kartverket/skipctl/pkg/ai"
	"github.com/spf13/cobra"
)

var (
	chatModel string
)

var chatCmd = &cobra.Command{
	Use:   "chat [question]",
	Short: "Chat with AI to help with manifests",
	Long: `Chat with AI about your manifests in natural language.

The AI assistant can help you:
- List and explore manifests
- Validate syntax and structure
- Render and preview changes
- Compare versions (diff)
- Format files
- Explain ArgoKit concepts

Examples:
  skipctl chat "what manifests do I have?"
  skipctl chat "validate all my manifests"
  skipctl chat "show me the diff for testdata/yaml/valid.yaml against HEAD"
  skipctl chat "is testdata/yaml/valid.yaml valid?"

Configuration:
Set your API key as an environment variable:
  export ANTHROPIC_API_KEY=sk-ant-...

Get your API key from: https://console.anthropic.com/

Interactive mode:
  skipctl chat
  > what manifests do I have?
  > validate them all
  > exit`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check for API key
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return fmt.Errorf(`No API key found. Please set:
  export ANTHROPIC_API_KEY=sk-ant-...
  
Get your API key from: https://console.anthropic.com/`)
		}

		agent := ai.NewAgent(apiKey, chatModel)

		// If question provided, run once and exit
		if len(args) > 0 {
			question := strings.Join(args, " ")
			return askQuestion(context.Background(), agent, question)
		}

		// Interactive mode
		return runInteractive(agent)
	},
}

func askQuestion(ctx context.Context, agent *ai.Agent, question string) error {
	fmt.Printf("🤔 Thinking...\n\n")

	response, err := agent.Ask(ctx, question)
	if err != nil {
		return fmt.Errorf("Chat request failed: %w", err)
	}

	fmt.Printf("Assistant: %s\n", response)
	return nil
}

func runInteractive(agent *ai.Agent) error {
	fmt.Println("💬 Skipctl Chat Assistant")
	fmt.Println("Ask me anything about your manifests. Type 'exit' to quit.\n")

	reader := bufio.NewReader(os.Stdin)
	ctx := context.Background()

	for {
		fmt.Print("You: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye! 👋")
			return nil
		}

		if err := askQuestion(ctx, agent, input); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		}
		fmt.Println()
	}
}

func init() {
	rootCmd.AddCommand(chatCmd)

	chatCmd.Flags().StringVar(&chatModel, "model", "", "Claude model to use (default: claude-3-haiku-20240307)")
}
