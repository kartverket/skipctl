
# skipctl

A simple client (and [server](./server.md)) to perform various network troubleshooting.

![skipctl](assets/skipctl-logo.png)

- [Installation](#installation)
- [Usage](#usage)
    - [Test](#test)
        - [Ping](#ping)
        - [Port probe](#port-probe)
    - [Manifests](#manifests)
        - [Render manifests](#render-manifests)
        - [Validate manifests](#validate-k8s-manifests)
- [Analytics & Privacy](#analytics--privacy)
## Installation

Download the [latest release](https://github.com/kartverket/skipctl/releases) or use the provided Docker image (mainly for running a server).

### Use Homebrew

```shell
brew tap kartverket/taps && \
brew install skipctl
```

## Usage

### Test
The various `test` commands will execute against an API server. Run `skipctl test ping --api-server=something` to get a list of supported API server names.
An API server represents a location that can run tests from their perspective. All communication with API servers is encrypted over TLS.

> :exclamation: Before issuing any `test` commands, be sure to be authenticated first (`gcloud auth application-default login`).


#### Ping

```shell
skipctl test ping --hostname=example.com --api-server=myApiServer
```

#### Port probe

```shell
skipctl test probe --hostname=example.com --port=1521 --api-server=myApiServer
```


### Manifests

The `manifests` command group contains commands for working with Skiperator manifests. All commands require a path pointing to a directory or file containing the manifest(s).
If a directory is specified, all commands in this group will recursively search the specified path for files with supported file formats. If no specific path is specified, the current working directory will be used.

**Supported file formats are:**
- `.jsonnet`
- `.libsonnet` (only for `manifests format`)
- `.yaml`
- `.yml`
- `kustomization.yaml`(only for `manifests render`)
- `kustomization.yml`(only for `manifests render`)


#### Render manifests
>[!note]
> When rendering kustomize, the renderer will ignore other `.jsonnet` and `.yaml` files in the same directory and subdirectories.

Compiles and renders a Skiperator `.jsonnet` or `.yaml` manifest in the specified directory and alerts if any errors are found.
```shell
skipctl manifests render <pathname>
```
#### Format manifests

Formats a `.jsonnet`, `.libsonnet` or `.yaml` manifest in the specified directory and alerts if any errors are found.
```shell
skipctl manifests format <pathname>
```

##### Format quick guide

Purpose: Rewrite manifests in-place to a canonical style (JSONNet / YAML). Recurses a directory or reads from stdin.

Inputs:
```
--path, -p <path>   Directory or file (defaults to CWD)
(stdin)             Use '-' as sole argument to read from standard input
```

Behavior:
- Only files with supported suffixes are touched (`.jsonnet`, `.yaml`, `.yml`).
- On error (parse / write) returns exit code 1 after logging.

Examples:
```shell
# Format everything under current directory
skipctl manifests format .

# Format a single file
skipctl manifests format ./app/manifest.yaml

# Format from stdin (outputs the formatted content to stdout)
cat manifest.yaml | skipctl manifests format -
```

Exit codes: 0 success / 1 error.
#### Diff manifests

Compares a Skiperator manifest against the currently deployed version in a Kubernetes cluster and shows the differences.

```shell
skipctl manifests diff --path <pathname> [flags]
```


##### Diff quick guide

Flags:
```
--ref <git-ref>        Git ref to diff against (default HEAD)
--diff-format <fmt>    pretty | patch | json (default pretty)
--verbosity <level>    full | chunk | minimal (auto-set if omitted)
--chunk-size <n>       Context lines for chunk (default 3)
--path, -p <path>      Files / directory to scan
```

Defaults (when --verbosity not provided): pretty->full, patch->chunk, json->full.

Examples:
```shell
skipctl manifests diff --path .
skipctl manifests diff --path . --diff-format json --verbosity minimal | jq '.diffs[] | select(.type!="Equals")'
```

Exit codes: 0 success / 1 error.



#### Validate K8s manifests

Validates a Skiperator manifest file (in either `.jsonnet` or `.yaml` format) against Skiperator's own custom schema definitions (skiperator.kartverket.no/v1alpha1)
**Supports the following resource types:**
- Application
- Routing
- SKIPJob

The API reference for these types can be found in the [SKIP docs](https://skip.kartverket.no/docs/applikasjon-utrulling/skiperator/api-docs).

Returns status code `1` if there are failures or `0` for successful validation

```shell
skipctl manifests validate <pathname>
```

#### Refactor manifests (Experimental)

The `refactor` command uses AI (Google Vertex AI with Gemini) to automatically refactor manifests. This command connects to a skipctl server instead of directly to GCP.

##### Setup

**Step 1: Set up Service Account Authentication**

A service account `skipctl@kv-spire-devex-ksde.iam.gserviceaccount.com` is configured for both server and client.

1. Create a service account key:
   ```shell
   gcloud iam service-accounts keys create ~/.config/gcloud/skipctl-key.json \
     --iam-account=skipctl@kv-spire-devex-ksde.iam.gserviceaccount.com
   ```

2. Set the environment variable (needed for both server and client):
   ```shell
   export GOOGLE_APPLICATION_CREDENTIALS="$HOME/.config/gcloud/skipctl-key.json"
   ```

> **Security Note:** 
> - Store service account keys securely in `~/.config/gcloud/` or use a secret management system
> - Never commit keys to version control
> - In production/GKE, use Workload Identity instead of key files (no GOOGLE_APPLICATION_CREDENTIALS needed)
> - Rotate keys regularly

**Step 2: Start the skipctl server**

```shell
# For local development (with self-signed TLS certificates)
# First, generate self-signed certificates:
mkdir -p ~/.config/skipctl
openssl req -x509 -newkey rsa:4096 -keyout ~/.config/skipctl/server-key.pem \
  -out ~/.config/skipctl/server-cert.pem -days 365 -nodes \
  -subj "/CN=localhost" -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

# Start the server with TLS
./skipctl serve --gcp-project-id=kv-spire-devex-ksde \
  --tls-cert=$HOME/.config/skipctl/server-cert.pem \
  --tls-key=$HOME/.config/skipctl/server-key.pem

# Or for production with proper certificates
./skipctl serve --gcp-project-id=kv-spire-devex-ksde \
  --tls-cert=/path/to/server-cert.pem \
  --tls-key=/path/to/server-key.pem \
  --gcp-location=europe-north1
```

> **Note:** Running the server without `--tls-cert` and `--tls-key` will start it in insecure mode (no encryption). This should only be used for local testing in trusted networks.

The server handles all communication with Vertex AI, so clients don't need GCP credentials.

**Step 3: Refactor your manifests**

```shell
# Refactor a single file
skipctl refactor app.jsonnet

# Provide additional files for context (libraries, shared configs, etc.)
skipctl refactor app.jsonnet lib/common.libsonnet shared/config.libsonnet

# Connect to a remote server
skipctl refactor --server=server.example.com:3514 app.jsonnet
```

The first file is the target to refactor. Additional files provide context to help the AI understand dependencies, shared functions, and configuration patterns. All files are sent to the AI, but only the first file is refactored.

Output is written to `vertexAI_output.libsonnet`.

##### Adding Custom Documentation Context

The AI service can be enhanced with custom documentation by adding files to the `docs/ai-context/` folder. The server automatically reads all `.md` and `.txt` files from this directory and includes them in the AI prompt.

**Why not use RAG (Retrieval-Augmented Generation)?**

While we initially considered using Google Cloud's Vertex AI Search with RAG for dynamic document retrieval, it requires **Google Cloud Enterprise Edition**, which is not available in our current setup. Instead, we use a simpler approach:

- All documentation files in `docs/ai-context/` are automatically concatenated and added to the AI prompt
- This provides the AI with custom context without requiring external services or enterprise features
- The approach is straightforward: just add your `.md` or `.txt` files to the folder

**How to add custom context:**

1. Create or add documentation files to `docs/ai-context/`:
   ```shell
   # Add custom documentation
   cat > docs/ai-context/my-patterns.md << 'EOF'
   # Custom Patterns
   
   ## Naming Conventions
   - Use kebab-case for application names
   - Prefix all resources with team identifier
   
   ## Best Practices
   - Always set resource limits
   - Use explicit namespace references
   EOF
   ```

2. The server will automatically load these files when processing refactor requests
3. You can add multiple files - they will all be included in the AI context

**Best practices:**
- Keep files focused on specific topics
- Use clear headings and structure
- Include examples where helpful
- Remove files that are no longer relevant

##### Security Considerations

**Network Security:**
- All client-server communication uses TLS encryption with system root CA certificates
- The server validates client authentication using Google Cloud ID tokens
- Only users from the `kartverket.no` organization can access the API
- See [server documentation](./server.md) for details on authentication

##### How It Works

1. Client sends manifest files to the skipctl server via gRPC
2. Server proxies the request to Vertex AI with Gemini models
3. Vertex AI Search provides grounding using ArgoKit v2 documentation and examples
4. Server returns the refactored content to the client
5. Refactored content is written to `vertexAI_output.libsonnet`

## Analytics & Privacy

`skipctl` collects anonymous usage analytics by **default** to help us understand how the tool is being used and improve the user experience.

### What We Collect

We collect the following **non-personal** information:
- **Command usage**: Which commands and subcommands are executed (e.g., `test ping`, `manifests render`)
- **Command arguments**: Arguments passed to commands (e.g., flags used)
- **Error information**: Whether a command succeeded or failed, and error messages if applicable. The error message may itself contain information about your system, e.g. file paths.
- **Environment context**:
  - Operating system (e.g., macOS, Linux, Windows)
  - System architecture (e.g., amd64, arm64)
  - Application version and git commit hash of skipctl
  - Environment type (local or CI)
- **Anonymous identifier**: A hashed machine-specific identifier that cannot be traced back to you.

### How We Collect Data

Analytics are collected through [PostHog](https://posthog.com/), a privacy-focused analytics platform, and sent to our self-hosted instance.

### How to Disable Analytics

You can disable analytics collection in two ways:

1. Set the `DO_NOT_TRACK` environment variable to `true`
2. Use the `--no-analytics` flag when running any `skipctl` command
