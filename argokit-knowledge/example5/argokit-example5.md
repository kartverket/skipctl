## This argokit v2:

local argokit = import '../../../argokit/v2/jsonnet/argokit.libsonnet';
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
   "apiVersion": "v1",
   "items": [
      {
         "apiVersion": "skiperator.kartverket.no/v1alpha1",
         "kind": "Application",
         "metadata": {
            "name": "foo-backend"
         },
         "spec": {
            "accessPolicy": {
               "outbound": {
                  "external": [
                     {
                        "host": "api.grafana.com"
                     }
                  ]
               }
            },
            "envFrom": [
               {
                  "secret": "open-ai-secrets"
               }
            ],
            "image": "foo-backend:1.2.3",
            "ingresses": [
               "devex.atgcp1-prod.kartverket-intern.cloud"
            ],
            "liveness": {
               "failureThreshold": 10,
               "initialDelay": 30,
               "path": "/health",
               "port": 8081,
               "timeout": 0
            },
            "port": 8080,
            "readiness": {
               "failureThreshold": 10,
               "initialDelay": 30,
               "path": "/health",
               "port": 8081,
               "timeout": 0
            }
         }
      },
      {
         "apiVersion": "external-secrets.io/v1",
         "kind": "ExternalSecret",
         "metadata": {
            "name": "open-ai-secrets"
         },
         "spec": {
            "data": [
               {
                  "remoteRef": {
                     "key": "openai-api-key",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "OPENAI_API_KEY"
               }
            ],
            "refreshInterval": "1h",
            "secretStoreRef": {
               "kind": "SecretStore",
               "name": "gsm"
            },
            "target": {
               "name": "open-ai-secrets"
            }
         }
      }
   ],
   "kind": "List"
}