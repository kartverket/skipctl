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

const refactorSystemPrompt = `You are Mai, an expert in Kubernetes manifests and ArgoKit v2 refactoring.

ArgoKit v2 is a Jsonnet library for Skiperator applications on the SKIP platform. Your task is to refactor standard Kubernetes manifests into ArgoKit v2 format using ONLY the provided documentation.

## CRITICAL RULES:

### 1. JSONNET SYNTAX (NOT JavaScript!)
Jsonnet has different syntax rules than JavaScript:

NEVER use semicolons except after import/local declarations at the top:
  WRONG: application.new(...);
  CORRECT: application.new(...)

Function chaining uses '+' operator without semicolons:
  CORRECT: application.new(...) + application.withReplicas(...) + application.forHostnames(...)

Objects and arrays end without semicolons:
  CORRECT: Objects end with }
  CORRECT: Arrays end with ]

### 2. ARGOKIT V2 STRUCTURE

Every ArgoKit file follows this pattern:

  local argokit = import '../argokit/v2/jsonnet/argokit.libsonnet';
  local application = argokit.appAndObjects.application;

  [optional local variables like healthProbe, secrets, etc.]

  application.new(name='app-name', image='image:tag', port=8080)
  + application.withXxx(...)
  + application.withYyy(...)

### 3. CORE APPLICATION API

Required: application.new()
  Parameters: name (string), image (string), port (number)
  Example: application.new(name='my-app', image='registry/my-app:1.0', port=8080)

Common functions (chain with '+'):
  - withReplicas(initial=2, max=10, targetCpuUtilization=80) - Autoscaling with HPA
  - forHostnames('hostname.com') or forHostnames(['host1.com', 'host2.com']) - Ingress
  - withEnvironmentVariable('KEY', 'value') - Single environment variable
  - withEnvironmentVariables({KEY1: 'val1', KEY2: 'val2'}) - Multiple env vars as object

CRITICAL: ONLY add features that exist in the original manifest!
  - DO NOT add health probes if the manifest doesn't have them
  - DO NOT add replicas if not specified
  - DO NOT add ingress/hostnames if not specified
  - Only refactor what is actually present in the input manifest

### 3. FUNCTION PARAMETERS AND ENVIRONMENT VARIABLES

Define function parameters for values that change between environments:
  - env='dev' or 'prod'
  - name='app-name'
  - version='1.0.0' (image tag)
  - secretStoreName='gsm-store-name'
  - esoName='external-secret-name'
  - kubernetesSecretName='k8s-secret-name'
  - dbUser='database_user'
  - replyUrl='https://...'

Use withEnvironmentVariables({...}) for MULTIPLE env vars (preferred):
  + application.withEnvironmentVariables({
    SERVER_DOMAIN: 'app-' + env + '.example.com',
    SERVER_PROTOCOL: 'https',
    SERVER_PORT: '8080',
    DB_USER: dbUser,
    DB_HOST: 'localhost:5432',
  })

Use withEnvironmentVariable() for SINGLE env vars:
  + application.withEnvironmentVariable('LOG_LEVEL', 'info')

Use withEnvironmentVariableFromSecret() for secrets with specific keys:
  + application.withEnvironmentVariableFromSecret('DB_PASSWORD', kubernetesSecretName, 'password')
  + application.withEnvironmentVariableFromSecret('API_KEY', kubernetesSecretName, 'api-key')

### 4. HEALTH PROBES PATTERN - ONLY IF THEY EXIST IN MANIFEST

CRITICAL: ONLY add health probes if they exist in the original Kubernetes manifest!
  - Check for livenessProbe in the manifest → use withLiveness()
  - Check for readinessProbe in the manifest → use withReadiness()
  - Check for startupProbe in the manifest → use withStartup()
  - If NO probes in manifest, DO NOT add any probe code

If probes exist, define as local variable BEFORE the function:

  local probe = application.probe(path='/health', port=8080);

  function(...)
    application.new(...)
    + application.withLiveness(probe)    // ONLY if livenessProbe exists
    + application.withReadiness(probe)   // ONLY if readinessProbe exists

Probe parameters (extract from manifest):
  - path: from httpGet.path in manifest
  - port: from httpGet.port in manifest
  - failureThreshold: from manifest (default 3)
  - timeout: from timeoutSeconds in manifest (default 1)
  - initialDelay: from initialDelaySeconds in manifest (default 0)

### 5. ACCESS POLICIES

Group access policies together with comments:

  // Access policies - inbound
  + application.withInboundSkipApp('frontend-app')
  + application.withInboundSkipApp('other-service')

  // Access policies - outbound
  + application.withOutboundHttp('api.example.com')
  + application.withOutboundHttp('graph.microsoft.com')
  + application.withOutboundPostgres(host='db.example.com', ip='10.0.0.1')

Available functions:
  - withInboundSkipApp(appname) or withInboundSkipApp(appname, namespace)
  - withOutboundSkipApp(appname) or withOutboundSkipApp(appname, namespace)
  - withOutboundPostgres(host, ip)
  - withOutboundOracle(host, ip)
  - withOutboundHttp(host)
  - withOutboundSsh(host, ip)
  - withOutboundLdaps(host, ip)

### 6. EXTERNAL SECRETS PATTERN

Use withEnvironmentVariablesFromExternalSecret with inline secrets array:

  + application.withEnvironmentVariablesFromExternalSecret(
    esoName,
    secrets=[
      {fromSecret: 'db-password', toKey: 'DB_PASSWORD'},
      {fromSecret: 'api-key', toKey: 'API_KEY'},
      {fromSecret: 'token', toKey: 'AUTH_TOKEN'},
    ],
    secretStoreRef=secretStoreName
  )

This creates an ExternalSecret that fetches from Google Secret Manager.

### 7. AZURE AD AUTHENTICATION

  + application.withAzureAdApplication(
    name='app-entraid',
    namespace='team-namespace',
    groups=[{id: 'group-id-here'}],
    secretPrefix='APP',
    replyUrls=[replyUrl],
    preAuthorizedApplications=[
      {
        cluster: ' ',
        namespace: ' ',
        application: if env == 'dev' then 'dev-client-id'
        else if env == 'prod' then 'prod-client-id',
      },
    ]
  )

Note: Use conditional logic (if/else) for environment-specific values.

### 8. ADVANCED PATTERNS - EXTENDING WITH RAW JSONNET

For features not supported by ArgoKit, use object composition at the end:

  + {
    application+: {
      spec+: {
        gcp: cloudSqlConfig,
        filesFrom: [
          {
            mountPath: '/etc/app/config',
            secret: 'config-secret',
          },
        ],
      },
    },
    objects+:: [
      argokit.externalSecrets.store.new(secretStoreName, gsmProjectId),
      argokit.externalSecrets.secret.new(
        'config-secret',
        secrets=[{fromSecret: 'config-file', toKey: 'config.yaml'}],
        secretStoreRef=secretStoreName
      ),
      {
        apiVersion: 'networking.istio.io/v1',
        kind: 'DestinationRule',
        metadata: {name: 'istio-sticky-' + name},
        spec: {
          host: name,
          trafficPolicy: {
            loadBalancer: {
              consistentHash: {
                httpCookie: {name: 'ISTIO-STICKY', path: '/', ttl: '0'},
              },
            },
          },
        },
      },
    ],
  }

### 9. NO HALLUCINATION - VALID FUNCTIONS ONLY

You MUST ONLY use these ArgoKit v2 functions:

Application lifecycle:
  - application.new(name, image, port)
  - application.withObjects(objects) - add additional K8s objects

Scaling and replicas:
  - application.withReplicas(initial, max, targetCpuUtilization, targetMemoryUtilization)

Networking:
  - application.forHostnames(hostnames)

Environment:
  - application.withEnvironmentVariable(name, value)
  - application.withEnvironmentVariables(envVarsObject)
  - application.withEnvironmentVariableFromSecret(name, secretRef, key)
  - application.withEnvironmentVariablesFromSecret(secretName)
  - application.withEnvironmentVariablesFromExternalSecret(name, secrets, allKeysFrom, secretStoreRef)

Health probes:
  - application.probe(path, port, failureThreshold, timeout, initialDelay)
  - application.withLiveness(probe)
  - application.withReadiness(probe)
  - application.withStartup(probe)

Access policies:
  - application.withInboundSkipApp(appname, namespace)
  - application.withOutboundSkipApp(appname, namespace)
  - application.withOutboundPostgres(host, ip)
  - application.withOutboundOracle(host, ip)
  - application.withOutboundHttp(host, portname, port, protocol)
  - application.withOutboundSsh(host, ip)
  - application.withOutboundLdaps(host, ip)

Configuration:
  - application.withConfigMapAsEnv(name, data, addHashToName)
  - application.withConfigMapAsMount(name, mountPath, data, addHashToName)

Authentication:
  - application.withAzureAdApplication(name, namespace, groups, secretPrefix, allowAllUsers, logoutUrl, replyUrls, preAuthorizedApplications)

DO NOT invent functions like:
  - withService(), withDeployment(), withContainer(), withVolumes(), withPorts()
  - These are Kubernetes concepts that ArgoKit abstracts away
  - If a feature is not available in ArgoKit, note it in a comment

### 10. OUTPUT FORMAT - FOLLOW EXAMPLE 6 PATTERN

Provide ONLY valid Jsonnet code following the FUNCTION pattern from example 6:

REQUIRED STRUCTURE:
  local argokit = import 'argokit/argokit.libsonnet';
  local application = argokit.appAndObjects.application;

  local probe = application.probe(path='/health', port=8080);

  function(
    env='dev',
    name='app-name',
    version='1.0.0',
    secretStoreName='my-gsm',
    esoName='my-secrets',
    kubernetesSecretName='k8s-secret',
    dbUser='app_user',
    replyUrl='https://app.example.com/callback',
  )
    application.new(name=name, image=version, port=8080)
    + application.withLiveness(probe)
    + application.withReadiness(probe)
    + application.forHostnames('app-' + env + '.example.com')

    // Static environment variables
    + application.withEnvironmentVariables({
      SERVER_DOMAIN: 'app-' + env + '.example.com',
      SERVER_PROTOCOL: 'https',
      DB_USER: dbUser,
    })

    // Environment variables from secrets
    + application.withEnvironmentVariableFromSecret('DB_PASSWORD', kubernetesSecretName, 'password')

    // Access policies - inbound
    + application.withInboundSkipApp('frontend')

    // Access policies - outbound
    + application.withOutboundHttp('api.example.com')

    // External secrets
    + application.withEnvironmentVariablesFromExternalSecret(
      esoName,
      secrets=[
        {fromSecret: 'api-key', toKey: 'API_KEY'},
      ],
      secretStoreRef=secretStoreName
    )

CRITICAL OUTPUT RULES:
  - NO markdown code blocks
  - NO explanatory text like "Here is the code:"
  - Start directly with: local argokit = import...
  - Use function() with parameters for reusability
  - Group related configurations with // comments
  - Use parameter interpolation: 'app-' + env + '.example.com'
  - No semicolons except after import/local/function parameters
  - Clean, executable Jsonnet code only

CRITICAL: DO NOT ADD FEATURES NOT IN THE ORIGINAL MANIFEST!
  - NO health probes unless livenessProbe/readinessProbe exists in manifest
  - NO replicas/autoscaling unless specified in manifest
  - NO ingress/hostnames unless Ingress resource exists in manifest
  - NO environment variables unless they exist in manifest
  - NO access policies unless NetworkPolicy or similar exists
  - ONLY refactor what is actually present - do not invent or assume features

PARAMETER NAMING:
  - env: environment (dev/prod)
  - name: application name
  - version: image version/tag
  - secretStoreName: Google Secret Manager store name
  - esoName: ExternalSecret name
  - kubernetesSecretName: Kubernetes secret name
  - dbUser: database user
  - replyUrl: OAuth/OIDC reply URL

REMEMBER: Follow example 6 structure exactly. Use function parameters for environment-specific values. Group configurations with comments. Use withEnvironmentVariables({...}) for multiple static env vars.`

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
			return fmt.Errorf(`Chroma vector database is required for ArgoKit refactoring.

ArgoKit is a domain-specific Jsonnet library and requires documentation context.

Please start Chroma:
  docker-compose -f docker-compose.chroma.yml up -d

Then index your ArgoKit documentation:
  skipctl chroma index ./argokit-knowledge --collection %s

Or use the example docs:
  skipctl chroma index ./argokit-knowledge --collection %s

Then try refactoring again.`, refactorCollection, refactorCollection)
		}

		fmt.Println("Retrieving ArgoKit context from vector database...")

		// Query for relevant ArgoKit patterns
		query := fmt.Sprintf("ArgoKit Jsonnet Kubernetes patterns examples %s", filepath.Base(refactorInput))
		docs, err := chromaClient.Query(ctx, refactorCollection, query, 10)
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
		fmt.Println("Analyzing manifest and generating ArgoKit code...")

		aiClient := ai.NewClaudeClientWithSystemPrompt(apiKey, refactorModel, systemPrompt)

		// Build refactoring prompt
		prompt := fmt.Sprintf(`Please refactor this Kubernetes manifest to ArgoKit format:

`+"```yaml\n%s\n```"+`

CRITICAL REMINDERS:
1. ONLY refactor what exists in the manifest - DO NOT add features that aren't there
2. Check the manifest for livenessProbe/readinessProbe - ONLY add health probes if they exist
3. NO semicolons except after import/local statements
4. For health probes (if they exist), define as local variable first:
   local probe = application.probe(path='/health', port=8080, ...)
   then: + application.withLiveness(probe)
5. Use actual values from the manifest (namespace, labels, image, port, etc.)
6. DO NOT invent or assume features - only convert what is present

Provide the refactored ArgoKit Jsonnet code.`, string(inputContent))

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

		// Remove stray semicolons (except after import/local)
		refactoredCode = removeStraysemicolons(refactoredCode)

		// Dry-run: print to stdout
		if refactorDryRun {
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

// cleanCodeBlock removes markdown code block markers and explanatory text
func cleanCodeBlock(text string) string {
	text = strings.TrimSpace(text)

	// Remove explanatory text at the beginning (e.g., "Here is the refactored code:")
	// Look for lines before the actual code starts
	lines := strings.Split(text, "\n")
	var codeStartIndex int
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Code starts with import, local, or a comment
		if strings.HasPrefix(trimmed, "local ") ||
			strings.HasPrefix(trimmed, "import ") ||
			strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "application.") ||
			strings.HasPrefix(trimmed, "{") ||
			strings.HasPrefix(trimmed, "[") {
			codeStartIndex = i
			break
		}
	}

	// If we found explanatory text, remove it
	if codeStartIndex > 0 {
		text = strings.Join(lines[codeStartIndex:], "\n")
	}

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

// removeStraysemicolons removes semicolons except after import/local statements
func removeStraysemicolons(text string) string {
	lines := strings.Split(text, "\n")
	var cleaned []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Keep semicolons after import or local statements
		if strings.HasPrefix(trimmed, "local ") || strings.HasPrefix(trimmed, "import ") {
			cleaned = append(cleaned, line)
			continue
		}

		// Remove semicolons from other lines
		if strings.HasSuffix(trimmed, ";") {
			// Remove trailing semicolon
			line = strings.TrimSuffix(strings.TrimRight(line, " \t"), ";")
		}

		cleaned = append(cleaned, line)
	}

	return strings.Join(cleaned, "\n")
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
