package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kartverket/skipctl/pkg/ai"
	"github.com/spf13/cobra"
)

var (
	chatModel            string
	chatDisableWrite     bool
	chatMaxFileSize      int64
	chatAllowAllPaths    bool
	chatSecurityProfile  string
	chatInstructions     string
	chatInstructionsFile string
)

var chatCmd = &cobra.Command{
	Use:   "chat [question]",
	Short: "Chat with AI to help with manifests",
	Long: `Chat with Mai about your manifests in natural language.

Mai can help you:
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

		// Load custom instructions
		var systemPrompt string
		if chatInstructionsFile != "" {
			content, err := os.ReadFile(chatInstructionsFile)
			if err != nil {
				return fmt.Errorf("failed to read instructions file: %w", err)
			}
			systemPrompt = string(content)
		} else if chatInstructions != "" {
			systemPrompt = chatInstructions
		}

		// Create security config based on flags
		var security *ai.SecurityConfig
		switch chatSecurityProfile {
		case "strict":
			security = ai.NewStrictSecurityConfig()
		case "relaxed":
			security = ai.NewRelaxedSecurityConfig()
		case "default", "":
			security = ai.DefaultSecurityConfig()
		default:
			return fmt.Errorf("unknown security profile: %s (use: default, strict, or relaxed)", chatSecurityProfile)
		}

		// Apply custom flags if provided
		if chatDisableWrite {
			security.DisableWrite = true
		}
		if chatMaxFileSize > 0 {
			security.MaxFileSize = chatMaxFileSize
		}
		if chatAllowAllPaths {
			security.WorkingDirOnly = false
		}

		// Create agent with security config and custom instructions
		var agent *ai.Agent
		if systemPrompt != "" {
			// Custom instructions provided
			agent = ai.NewAgentWithOptions(apiKey, chatModel, systemPrompt, security, 20, time.Minute)
		} else if chatSecurityProfile != "" || chatDisableWrite || chatMaxFileSize > 0 || chatAllowAllPaths {
			// Custom security config
			rateLimiter := ai.NewRateLimiter(20, time.Minute)
			agent = ai.NewAgentWithSecurity(apiKey, chatModel, security, rateLimiter)
		} else {
			// Default config
			agent = ai.NewAgent(apiKey, chatModel)
		}

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

	fmt.Printf("Mai: %s\n", response)
	return nil
}

func runInteractive(agent *ai.Agent) error {
	fmt.Println("💬 Mai - Skipctl Chat")
	fmt.Println("Ask me anything about your manifests. Type 'exit' to quit.")

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

	// Instruction flags
	chatCmd.Flags().StringVar(&chatInstructions, "instructions", "", "Custom instructions for Mai (overrides default behavior)")
	chatCmd.Flags().StringVar(&chatInstructionsFile, "instructions-file", "", "Load custom instructions from a file")

	// Security flags
	chatCmd.Flags().StringVar(&chatSecurityProfile, "security", "default", "Security profile: default, strict, or relaxed")
	chatCmd.Flags().BoolVar(&chatDisableWrite, "read-only", false, "Disable write operations (format)")
	chatCmd.Flags().Int64Var(&chatMaxFileSize, "max-file-size", 0, "Maximum file size in bytes (0 = use profile default)")
	chatCmd.Flags().BoolVar(&chatAllowAllPaths, "allow-all-paths", false, "Allow access outside working directory")
}
