# MCP Usage Examples

Here are real-world examples of how to use Skipctl with AI assistants through the MCP protocol.

## Example Conversations

### 1. Understanding Existing Manifests

**You:** "What manifests do I have in this project?"

**Claude (using `list_manifests`):** "You have 15 manifest files in your project:
- manifests/prod/app.yaml
- manifests/prod/database.yaml
- manifests/staging/app.yaml
..."

**You:** "Show me what the prod app manifest will deploy"

**Claude (using `render_manifest`):** "Here's the rendered production app manifest:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
...
```"

### 2. Validating Changes

**You:** "I just modified manifests/prod/app.yaml. Can you check if it's valid?"

**Claude (using `validate_manifest`):** "✅ The manifest is valid and can be parsed correctly."

**You:** "What will change compared to what's deployed in production?"

**Claude (using `diff_manifest` with ref='main'):** "Here are the changes:
- Increased replicas from 2 to 5
- Updated image tag from v1.2.3 to v1.2.4
- Added a new environment variable: FEATURE_FLAG_X=true"

### 3. Formatting and Cleanup

**You:** "Format all my YAML files"

**Claude (using `list_manifests` + `format_manifest`):** "I'll format your manifest files:
✅ Formatted manifests/prod/app.yaml
✅ Formatted manifests/prod/database.yaml
✅ Formatted manifests/staging/app.yaml
..."

### 4. Learning ArgoKit

**You:** "How do I add a Redis cache to my application?"

**Claude:** "I'll help you add Redis. First, let me see your current manifest structure..."

**Claude (using `list_manifests` and `render_manifest`):**

**Claude:** "Based on your current setup, you should create a new file `manifests/prod/redis.yaml` with this ArgoKit configuration:

```yaml
apiVersion: skiperator.kartverket.no/v1alpha1
kind: Application
metadata:
  name: myapp-redis
  namespace: myapp-prod
spec:
  image: redis:7-alpine
  port: 6379
  replicas:
    min: 1
    max: 1
  resources:
    requests:
      memory: "256Mi"
      cpu: "100m"
    limits:
      memory: "2Gi"
      cpu: "500m"
```

Would you like me to validate this configuration?"

### 5. Debugging Issues

**You:** "My deployment is failing. Can you check my manifests?"

**Claude (using `validate_manifest` on all files):** "I found an issue in manifests/prod/app.yaml:
❌ Manifest failed to render: invalid port configuration

Let me look at the file..."

**Claude (using `render_manifest`):** "The problem is on line 15 - the port must be a number, but you have it as a string '8080'. Change it to:
```yaml
port: 8080
```"

### 6. Pre-deployment Review

**You:** "I'm about to deploy to production. Can you review my changes?"

**Claude (using `diff_manifest` against 'main'):** "Let me check your changes against the main branch:

**Changes detected:**

1. **manifests/prod/app.yaml**
   - Image updated: v2.1.0 → v2.2.0
   - New environment variable: DATABASE_POOL_SIZE=20
   
2. **manifests/prod/database.yaml**
   - Memory limit increased: 4Gi → 8Gi

**Validation:** ✅ All manifests are valid

**Recommendations:**
- The memory increase for the database is significant. Make sure this aligns with your infrastructure capacity.
- Consider testing v2.2.0 in staging first if you haven't already.

Ready to deploy?"

### 7. Migration Help

**You:** "I need to convert my old Kubernetes YAML to ArgoKit format"

**Claude (using `render_manifest` to see current format):** "I can help you migrate. Let me see your current deployment..."

**Claude:** "Here's your current deployment converted to ArgoKit format:

```yaml
apiVersion: skiperator.kartverket.no/v1alpha1
kind: Application
metadata:
  name: myapp
  namespace: production
spec:
  image: "myorg/myapp:v1.0.0"
  port: 8080
  replicas:
    min: 2
    max: 10
  resources:
    requests:
      cpu: "200m"
      memory: "512Mi"
    limits:
      cpu: "1000m"
      memory: "1Gi"
  ingress:
    enabled: true
```

This ArgoKit manifest replaces your Deployment, Service, and HorizontalPodAutoscaler with a single, simpler configuration."

## Complex Workflows

### Bulk Operations

**You:** "Check all manifests in the staging directory for issues"

**Claude:**
1. Uses `list_manifests` to find all files in staging/
2. Uses `validate_manifest` on each file
3. Uses `render_manifest` to verify rendering
4. Reports results in a table format

### Environment Comparison

**You:** "What's different between my prod and staging configurations?"

**Claude:**
1. Uses `list_manifests` for both directories
2. Uses `render_manifest` to compare equivalent files
3. Highlights differences in:
   - Resource allocations
   - Replica counts
   - Environment variables
   - Image tags

### Interactive Debugging

**You:** "My Jsonnet manifest has an error"

**Claude (using `validate_manifest`):** "Error on line 42: undefined variable 'namespace'"

**You:** "How do I fix that?"

**Claude:** "You need to define the namespace variable. In Jsonnet, you can either:

1. Add it as a parameter:
```jsonnet
function(namespace='default') {
  // your manifest
}
```

2. Import it from a library:
```jsonnet
local config = import 'config.libsonnet';
local namespace = config.namespace;
```

Would you like me to validate the file again after you make the change?"

## Tips for Working with AI + Skipctl

1. **Be specific about files**: "Check manifests/prod/app.yaml" vs "check my manifest"
2. **Ask for validation after changes**: Always validate after AI suggests modifications
3. **Use git refs for comparison**: "Compare against main branch" or "show diff from v1.0.0"
4. **Request explanations**: "Explain what this manifest will deploy"
5. **Iterate**: Start with validation, then rendering, then deployment prep

## Integration Patterns

### CI/CD Integration

You can use the MCP server in CI/CD by scripting the JSON-RPC calls:

```bash
# Validate all manifests
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"list_manifests","arguments":{}},"id":1}' | skipctl mcp

# Then validate each one programmatically
```

### Local Development Workflow

1. Edit manifest files in your editor
2. Ask Claude to validate changes
3. Review diffs against main branch
4. Format files
5. Commit with AI-generated commit messages

### Learning and Documentation

Use Claude to:
- Generate documentation for your manifests
- Create runbooks for deployment procedures
- Explain complex Jsonnet logic
- Suggest ArgoKit best practices
