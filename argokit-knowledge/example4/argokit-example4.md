## This argokit v2:

local argokit = import '../../argokit/v2/jsonnet/argokit.libsonnet';
local application = argokit.appAndObjects.application;

local healthProbe = application.probe(
  path='/health',
  port=8081,
  failureThreshold=10,
  timeout=0,
  initialDelay=30
);

application.new(
  name='foo-backend',
  image='foo-backend:1.2.3',
  port=8080
)

+ application.withEnvironmentVariables({
  WAS_SCANNER_NAME: 'my-scanner-name',
  WAS_SCANNER_LOG_FILE: '/tmp/scanner.log',
})

+ application.withLiveness(healthProbe)
+ application.withReadiness(healthProbe)
+ application.forHostnames('devex.atgcp1-' + 'prod' + '.kartverket-intern.cloud')
+ application.withOutboundHttp('api.grafana.com')

+ application.withEnvironmentVariablesFromExternalSecret(
  name='open-ai-secrets',
  secrets=[
    {
      fromSecret: 'openai-api-key',
      toKey: 'OPENAI_API_KEY',
    },
  ],
)

## Renders to this Kubernetes Application manifest:

{
  apiVersion: 'v1',
  items: [
    {
      apiVersion: 'skiperator.kartverket.no/v1alpha1',
      kind: 'Application',
      metadata: {
        name: 'foo-backend',
      },
      spec: {
        image: 'foo-backend:1.2.3',
        port: 8080,
        env: [
          {
            key: 'WAS_SCANNER_LOG_FILE',
            value: '/tmp/scanner.log',
          },
          {
            key: 'WAS_SCANNER_NAME',
            value: 'my-scanner-name',
          },
        ],
        envFrom: [
          {
            secret: 'open-ai-secrets',
          },
        ],
        liveness: {
          path: '/health',
          port: 8081,
          failureThreshold: 10,
          initialDelay: 30,
          timeout: 0,
        },
        readiness: {
          path: '/health',
          port: 8081,
          failureThreshold: 10,
          initialDelay: 30,
          timeout: 0,
        },
        ingresses: ['devex.atgcp1-' + 'prod' + '.kartverket-intern.cloud'],
        accessPolicy: {
          outbound: {
            external: [
              {
                host: 'api.grafana.com',
              },
            ],
          },
        },
      },
    },
    {
      apiVersion: 'external-secrets.io/v1',
      kind: 'ExternalSecret',
      metadata: {
        name: 'open-ai-secrets',
      },
      spec: {
        refreshInterval: '1h',
        secretStoreRef: {
          kind: 'SecretStore',
          name: 'gsm',
        },
        target: {
          name: 'open-ai-secrets',
        },
        data: [
          {
            secretKey: 'OPENAI_API_KEY',
            remoteRef: {
              key: 'openai-api-key',
              metadataPolicy: 'None',
            },
          },
        ],
      },
    },
  ],
  kind: 'List',
}
