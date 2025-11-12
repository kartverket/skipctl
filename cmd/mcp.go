package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/kartverket/skipctl/pkg/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server for AI integration",
	Long: `Start a Model Context Protocol (MCP) server that allows AI assistants
like Claude to interact with Skipctl's manifest tools.

The MCP server provides tools for:
- Rendering manifests (Jsonnet, YAML, Kustomize)
- Comparing manifest changes (diff)
- Validating manifest syntax
- Formatting manifest files
- Listing available manifests

Example usage with Claude Desktop:
Add to ~/Library/Application Support/Claude/claude_desktop_config.json:
{
  "mcpServers": {
    "skipctl": {
      "command": "/path/to/skipctl",
      "args": ["mcp"]
    }
  }
}

Then restart Claude Desktop and ask questions like:
"Show me the diff for my manifest against main branch"
"Validate all manifests in the current directory"
"Render the production deployment manifest"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle signals for graceful shutdown
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigCh
			cancel()
		}()

		server := mcp.NewServer(os.Stdin, os.Stdout)
		return server.Serve(ctx)
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
