# Skipctl Refactor - AI-Powered ArgoKit Migration

Automatically refactor Kubernetes manifests to ArgoKit format using Claude AI.

## What is this?

A command that takes your existing Kubernetes YAML/JSON manifests and converts them to ArgoKit's Jsonnet-based format with best practices applied automatically.

**What it does:**
- 🔄 Converts Kubernetes YAML to ArgoKit Jsonnet
- ✅ Applies best practices (resource limits, health checks, security)
- 📚 Optionally uses your ArgoKit docs/examples for context (vector DB)
- 🚀 Maintains all functionality from the original manifest

## Quick Start

### 1. Set API Key

```bash
export ANTHROPIC_API_KEY=sk-ant-...
```

Add to `~/.zshrc` or `~/.bashrc` for persistence:
```bash
echo 'export ANTHROPIC_API_KEY=sk-ant-...' >> ~/.zshrc
source ~/.zshrc
```

### 2. Refactor a Manifest

```bash
# Basic refactoring
skipctl refactor deployment.yaml

# Custom output path
skipctl refactor deployment.yaml -o argokit/app.jsonnet

# Preview without writing
skipctl refactor deployment.yaml --dry-run
```

## Usage Examples

### Basic Refactoring

```bash
# Refactor a deployment
skipctl refactor k8s/deployment.yaml

# Output: k8s/deployment.argokit.jsonnet
```

### With Custom Output

```bash
# Specify output location
skipctl refactor manifests/app.yaml -o argokit/applications/app.jsonnet

# Create parent directories automatically
skipctl refactor old/service.yaml -o new/structure/service.jsonnet
```

### Preview Mode

```bash
# See the output without writing
skipctl refactor deployment.yaml --dry-run

# Pipe to file manually
skipctl refactor deployment.yaml --dry-run > app.jsonnet
```

### With Custom Instructions

```bash
# Focus on specific aspects
skipctl refactor app.yaml --instructions "Focus on security best practices"

# Add specific requirements
skipctl refactor app.yaml --instructions "Use minimal resource limits for development"

# Multiple concerns
skipctl refactor app.yaml --instructions "Add monitoring labels and optimize for cost"
```

### Using Vector DB Context

```bash
# Provide ArgoKit documentation and examples
skipctl refactor app.yaml --vector-db ./argokit-docs

# Use project-specific examples
skipctl refactor app.yaml --vector-db ./examples/argokit
```

The vector DB path should contain:
- `.md` files with documentation
- `.jsonnet` files with ArgoKit examples
- Your own ArgoKit patterns and templates

## How It Works

1. **Reads** your Kubernetes manifest (YAML/JSON)
2. **Analyzes** the structure (Deployments, Services, Ingress, etc.)
3. **Converts** to ArgoKit Jsonnet format
4. **Applies** best practices:
   - Resource limits and requests
   - Health checks (readiness, liveness)
   - Security contexts
   - Proper labels and annotations
   - Monitoring integration
5. **Writes** the refactored Jsonnet code

### What Gets Added

The AI automatically enhances your manifests:

- **Resource Management**: CPU/memory limits if missing
- **Health Checks**: Readiness and liveness probes
- **Security**: SecurityContext, non-root users
- **Monitoring**: Prometheus labels
- **Best Practices**: Proper naming, namespaces, selectors

### What Gets Preserved

Everything from your original manifest:
- Container images and versions
- Environment variables
- Volumes and mounts
- Services and ports
- Ingress rules
- ConfigMaps/Secrets references

## Advanced Usage

### Batch Refactoring

```bash
# Refactor multiple files
for file in k8s/*.yaml; do
  skipctl refactor "$file" -o "argokit/$(basename "$file" .yaml).jsonnet"
done
```

### With Different Models

```bash
# Use a more powerful model for complex manifests
skipctl refactor complex-app.yaml --model claude-3-5-sonnet-20241022

# Cost-effective for simple manifests (default)
skipctl refactor simple-app.yaml --model claude-3-haiku-20240307
```

### Integrate with Git Workflow

```bash
# Create a new branch for refactoring
git checkout -b argokit-migration

# Refactor all manifests
mkdir -p argokit
for manifest in k8s/*.yaml; do
  skipctl refactor "$manifest" -o "argokit/$(basename "$manifest" .yaml).jsonnet"
done

# Review changes
git diff --no-index k8s/ argokit/

# Commit
git add argokit/
git commit -m "Refactor manifests to ArgoKit format"
```

## Vector Database Setup

To provide ArgoKit context for better refactoring:

### 1. Prepare Documentation

```bash
mkdir argokit-context
cd argokit-context

# Add ArgoKit docs
cp -r /path/to/argokit/docs/*.md .

# Add your examples
cp -r ../examples/*.jsonnet .

# Add templates
cp -r ../templates/*.jsonnet .
```

### 2. Use with Refactor

```bash
skipctl refactor app.yaml --vector-db ./argokit-context
```

The AI will:
- Read documentation and examples
- Apply patterns from your codebase
- Use consistent naming conventions
- Follow your team's practices

### Future: Automatic Vector DB

**Coming soon**: Integration with vector databases (ChromaDB, Pinecone) for:
- Semantic search across documentation
- Automatic pattern detection
- Team knowledge base
- Version-specific guidance

## Cost Estimates

Claude API pricing (approximate):

| Operation | Input Tokens | Output Tokens | Cost |
|-----------|--------------|---------------|------|
| Simple manifest | ~1,000 | ~500 | $0.003 |
| Complex manifest | ~3,000 | ~1,500 | $0.009 |
| With vector context | ~8,000 | ~2,000 | $0.018 |

**Tips to reduce costs:**
- Use `--dry-run` to preview before writing
- Start with simple manifests to test
- Use default Haiku model (fast & cheap)
- Provide clear `--instructions` to reduce iterations

## Troubleshooting

### "No API key found"

```bash
# Set the API key
export ANTHROPIC_API_KEY=sk-ant-...

# Verify it's set
echo $ANTHROPIC_API_KEY
```

### "input file not found"

Use absolute or relative paths:
```bash
# Relative
skipctl refactor ./k8s/deployment.yaml

# Absolute
skipctl refactor /full/path/to/deployment.yaml
```

### "Failed to read input file"

Check file permissions:
```bash
ls -l deployment.yaml
chmod 644 deployment.yaml
```

### Generated Code Has Issues

1. **Try with instructions:**
   ```bash
   skipctl refactor app.yaml --instructions "Simplify structure, focus on core functionality"
   ```

2. **Use vector DB context:**
   ```bash
   skipctl refactor app.yaml --vector-db ./examples
   ```

3. **Try a better model:**
   ```bash
   skipctl refactor app.yaml --model claude-3-5-sonnet-20241022
   ```

### Output Doesn't Match Expectations

The AI bases output on:
- Standard ArgoKit patterns
- Kubernetes best practices
- Your custom instructions
- Vector DB context (if provided)

Provide specific instructions:
```bash
skipctl refactor app.yaml --instructions "Use specific pattern from docs/example.jsonnet"
```

## Best Practices

### 1. Start with Dry-Run

Always preview first:
```bash
skipctl refactor app.yaml --dry-run | less
```

### 2. Review Generated Code

Don't blindly trust AI output:
- Check resource limits are appropriate
- Verify health check paths
- Ensure secrets aren't hardcoded
- Test in development first

### 3. Use Version Control

```bash
git checkout -b argokit-refactor
skipctl refactor app.yaml
git diff  # Review changes
git commit
```

### 4. Provide Context

Better context = better output:
```bash
# Good
skipctl refactor app.yaml \
  --vector-db ./argokit-docs \
  --instructions "Follow our standard monitoring pattern"

# Better
skipctl refactor app.yaml \
  --vector-db ./argokit-docs \
  --instructions "Use pattern from examples/web-app.jsonnet, add Prometheus metrics"
```

### 5. Iterate

Refine the output:
```bash
# First pass
skipctl refactor app.yaml -o app.v1.jsonnet

# Refine with feedback
skipctl refactor app.yaml -o app.v2.jsonnet \
  --instructions "Reduce memory limits, add init container"
```

## Examples

### Simple Deployment

**Input** (`deployment.yaml`):
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: app
        image: myapp:1.0
        ports:
        - containerPort: 8080
```

**Command**:
```bash
skipctl refactor deployment.yaml --dry-run
```

**Output** (`deployment.argokit.jsonnet`):
```jsonnet
local ArgoKit = import 'argo-kit/kube.libsonnet';

local app = ArgoKit.Application {
  metadata: {
    name: 'myapp',
    namespace: 'default',
  },
  spec: {
    replicas: 3,
    containers: [
      {
        name: 'app',
        image: 'myapp:1.0',
        ports: [{ containerPort: 8080 }],
        resources: {
          requests: { cpu: '100m', memory: '128Mi' },
          limits: { cpu: '500m', memory: '512Mi' },
        },
        readinessProbe: {
          httpGet: { path: '/health', port: 8080 },
        },
      },
    ],
  },
};

app
```

### With Service and Ingress

**Input** (`app.yaml`):
```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.21
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: web
spec:
  ports:
  - port: 80
  selector:
    app: web
```

**Command**:
```bash
skipctl refactor app.yaml --instructions "Add TLS ingress"
```

## FAQ

### Q: Does it work with Helm charts?

Not directly. Convert Helm to YAML first:
```bash
helm template myapp ./chart > app.yaml
skipctl refactor app.yaml
```

### Q: Can I refactor back to vanilla K8s?

The generated Jsonnet can be rendered:
```bash
skipctl render app.argokit.jsonnet > app.yaml
```

### Q: Does it support Kustomize?

Yes! Render first:
```bash
kubectl kustomize ./overlay > app.yaml
skipctl refactor app.yaml
```

### Q: What about secrets?

Secrets are converted with references preserved. Never paste actual secret values in manifests.

### Q: Can I customize ArgoKit patterns?

Yes! Put your patterns in the vector DB directory:
```bash
skipctl refactor app.yaml --vector-db ./my-argokit-patterns
```

### Q: Does it handle StatefulSets, Jobs, CronJobs?

Yes, all Kubernetes resource types are supported.

## Security Notes

⚠️ **Important:**
- Manifest content is sent to Claude API (Anthropic)
- Don't refactor manifests with embedded secrets
- API calls are over HTTPS (encrypted)
- Anthropic doesn't train on your data

**Best practices:**
- Use secret references, not values
- Review output before committing
- Test in development first
- Use `--dry-run` for sensitive manifests

## Next Steps

1. **Try it out:**
   ```bash
   skipctl refactor testdata/yaml/valid.yaml --dry-run
   ```

2. **Set up vector DB** with your ArgoKit docs

3. **Refactor a real manifest:**
   ```bash
   skipctl refactor production/app.yaml -o argokit/app.jsonnet
   ```

4. **Review and test** the generated code

5. **Iterate** with custom instructions as needed

## Related Commands

- `skipctl chat` - Interactive AI assistant for manifests
- `skipctl render` - Render Jsonnet to YAML
- `skipctl validate` - Validate manifests
- `skipctl diff` - Compare manifest versions

## Feedback

Found an issue or have suggestions? The refactor command learns from:
- Your custom instructions
- Vector DB context you provide
- The ArgoKit patterns in your codebase

Provide better context for better results!
