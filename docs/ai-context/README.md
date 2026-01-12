# AI Context Directory

This directory contains additional documentation that is automatically loaded and included in AI refactoring prompts.

## Purpose

When you run `skipctl refactor`, the server reads all `.md` and `.txt` files from this directory and adds them to the prompt sent to Vertex AI. This allows you to provide custom context like:

- Organization-specific patterns and conventions
- Best practices for your team
- Common code patterns and examples
- Migration guides
- API documentation
- Architecture decisions

## How It Works

1. All `.md` and `.txt` files in this directory are automatically scanned
2. Files are concatenated with clear separators showing which file each section comes from
3. The combined content is appended to every AI refactoring request
4. The AI uses this context to better understand your requirements and produce more relevant output

## Adding Documentation

Simply add or edit `.md` or `.txt` files in this directory. The server will pick them up on the next refactor request. No restart required.

### Example

```bash
# Add your own patterns
cat > docs/ai-context/team-conventions.md << 'EOF'
# Team Conventions

## Naming Standards
- Use kebab-case for all application names
- Prefix resources with team identifier: `myteam-`

## Required Configurations
- Always include resource limits
- Always add liveness and readiness probes
EOF
```

## Why Not RAG?

We initially considered using Retrieval-Augmented Generation (RAG) with Vertex AI Search, but it requires **Google Cloud Enterprise Edition**. Instead, we use this simpler folder-based approach that doesn't require external services or enterprise features.

## Current Files

- **Argokit-README.md** - ArgoKit v2 documentation
- **example1.md** - Real-world ArgoKit migration examples
- **example2.md** - Additional migration patterns

Feel free to add more files or modify existing ones to better match your needs!
