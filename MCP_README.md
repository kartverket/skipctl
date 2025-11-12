# Skipctl MCP Server

Skipctl now includes a Model Context Protocol (MCP) server that allows AI assistants like Claude to interact directly with Skipctl's manifest tools.

## What is MCP?

Model Context Protocol (MCP) is an open standard that enables AI assistants to securely connect to external tools and data sources. By running Skipctl as an MCP server, AI assistants can:

- Render manifests (Jsonnet, YAML, Kustomize)
- Compare manifest changes (diff)
- Validate manifest syntax
- Format manifest files
- List available manifests

## Available Tools

### `render_manifest`
Render a manifest file to see the final YAML output.

**Parameters:**
- `file` (required): Path to manifest file

**Example:**
```
"Show me the rendered output of manifests/prod/myapp.yaml"
```

### `diff_manifest`
Compare manifest changes between current state and a git reference.

**Parameters:**
- `file` (required): Path to manifest file
- `ref` (optional): Git reference to compare against (default: HEAD)

**Example:**
```
"What changes will occur if I apply manifests/prod/myapp.yaml compared to main branch?"
```

### `validate_manifest`
Validate manifest syntax and structure.

**Parameters:**
- `file` (required): Path to manifest file

**Example:**
```
"Check if my manifest is valid"
```

### `format_manifest`
Format a manifest file according to standard conventions.

**Parameters:**
- `file` (required): Path to manifest file

**Example:**
```
"Format my manifest file"
```

### `list_manifests`
List all manifest files in a directory.

**Parameters:**
- `path` (optional): Directory path to search (default: current directory)

**Example:**
```
"Show me all manifest files in the project"
```

## Setup with Claude Desktop

1. Build Skipctl:
   ```bash
   go build -o skipctl .
   ```

2. Find the full path to your skipctl binary:
   ```bash
   pwd # Note this path
   ```

3. Edit Claude Desktop config file:
   ```bash
   # macOS
   code ~/Library/Application\ Support/Claude/claude_desktop_config.json
   
   # Linux
   code ~/.config/Claude/claude_desktop_config.json
   ```

4. Add Skipctl as an MCP server:
   ```json
   {
     "mcpServers": {
       "skipctl": {
         "command": "/full/path/to/skipctl",
         "args": ["mcp"]
       }
     }
   }
   ```

5. Restart Claude Desktop

6. Look for the 🔌 icon in Claude Desktop to verify the connection

## Example Conversations

Once configured, you can ask Claude questions like:

- "Show me the diff for manifests/prod/myapp.yaml against main branch"
- "Validate all manifests in the current directory"
- "Render the production deployment manifest"
- "List all manifest files in this project"
- "Format my Jsonnet manifests"
- "What will change if I deploy this manifest?"

## Testing MCP Server Manually

You can test the MCP server directly:

```bash
# Initialize
echo '{"jsonrpc":"2.0","method":"initialize","id":1}' | ./skipctl mcp

# List available tools
echo '{"jsonrpc":"2.0","method":"tools/list","id":2}' | ./skipctl mcp

# Render a manifest
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"render_manifest","arguments":{"file":"testdata/yaml/valid.yaml"}},"id":3}' | ./skipctl mcp

# Validate a manifest
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"validate_manifest","arguments":{"file":"testdata/yaml/valid.yaml"}},"id":4}' | ./skipctl mcp
```

## How It Works

The MCP server:
1. Listens on stdin for JSON-RPC 2.0 messages
2. Responds on stdout with JSON-RPC 2.0 responses
3. Uses existing Skipctl packages (manifest, diff, etc.)
4. Provides structured output that AI assistants can understand

## Benefits

- **Lower barrier to entry**: AI can help users learn ArgoKit syntax
- **Faster workflows**: Natural language commands instead of remembering CLI flags
- **Error prevention**: AI can validate before applying changes
- **Documentation**: AI provides context-aware help
- **Exploration**: AI can discover and explain existing manifests

## Architecture

```
┌─────────────────┐
│  Claude Desktop │
│   (AI Client)   │
└────────┬────────┘
         │ MCP Protocol (stdin/stdout)
         │ JSON-RPC 2.0
         ▼
┌─────────────────┐
│  Skipctl MCP    │
│     Server      │
├─────────────────┤
│ - render        │
│ - diff          │
│ - validate      │
│ - format        │
│ - list          │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Skipctl Pkg   │
│  - manifest     │
│  - diff         │
│  - git          │
└─────────────────┘
```

## Troubleshooting

### Connection issues
- Verify the path to skipctl binary is correct and absolute
- Check that skipctl has execute permissions: `chmod +x skipctl`
- Look at Claude Desktop logs for errors

### Tool not working
- Ensure you're in a directory with manifest files
- Check that file paths are correct (relative or absolute)
- For diff operations, ensure you're in a git repository

## Future Enhancements

Potential additions:
- Apply manifests directly (with confirmation)
- Create new manifests from templates
- Explain existing manifests
- Suggest improvements based on best practices
- Integration with ArgoCD to check deployment status
