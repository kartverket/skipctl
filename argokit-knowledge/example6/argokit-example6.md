## This argokit v2:

local argokit = import '../argokit/v2/jsonnet/argokit.libsonnet';
local application = argokit.appAndObjects.application;

local probe = application.probe(path='/health', port=8080);

function(
  env='dev',
  name='demo-backend',
  version='1',
  gsmProjectId='very-cool-version',
  secretStoreName='devex-demo-gsm',
  esoName='devex-demo-secrets',
  kubernetesSecretEntraIdSecretName='entraid-secret',
  cloudSqlConfig={
    instanceConnectionName: 'project:region:instance',
    databaseUser: 'demo_user',
    databaseName: 'demo_db',
    cloudSqlProxyImage: 'gcr.io/cloudsql-docker/gce-proxy:1.33.3',
  },
  dbUser='demo_user',
  replyUrl='https://demo.atgcp1-dev.host.com/callback',
)
  application.new(name=name, image=version, port=8080)
  + application.withLiveness(probe)
  + application.withReadiness(probe)
  + application.forHostnames('backend.atgcp1-' + env + 'host.com')

  // Static environment variables
  + application.withEnvironmentVariables({
    RR_SERVER_DOMAIN: 'demo.atgcp1-' + env + '.host.com',
    RR_SERVER_PROTOCOL: 'https',
    RR_SERVER_HTTP_PORT: '8080',
    RR_SERVER_ROOT_URL: 'https://demo.atgcp1-' + env + '.host.com',
    FRONTEND_URL_HOST: 'demo.atgcp1-' + env + '.host.com',
    TENANT_ID: 'ogaboga-abcd-1234-gagaga-lgtm-123',
    AUTH_PROVIDER_URL: 'https://demo.atgcp1-' + env + '.host.com/callback',
    DB_NAME: dbUser,
    DB_PASSWORD: '',
    FRISK_FRONTEND_URL_HOST: 'https://frisk.atgcp1-' + env + '.host.com',
    RR_OAUTH_TENANT_ID: 'f9f9f-abcd-1234-gagaga-lgtm-123',
    RR_DATABASE_HOST: 'localhost:5432',
    RR_DATABASE_NAME: 'demo',
    RR_DATABASE_USER: dbUser,
    RR_DATABASE_PASSWORD: '',
    RR_SERVER_ALLOWED_ORIGINS: 'frisk.atgcp1-' + env + '.host.com',
  })

  // Environment variables from secrets with specific keys
  + application.withEnvironmentVariableFromSecret('RR_OAUTH_CLIENT_ID', kubernetesSecretEntraIdSecretName, 'RR_APP_CLIENT_ID')
  + application.withEnvironmentVariableFromSecret('RR_OAUTH_CLIENT_SECRET', kubernetesSecretEntraIdSecretName, 'RR_APP_CLIENT_SECRET')

  // Access policies - inbound
  + application.withInboundSkipApp('demo-frontend')
  + application.withInboundSkipApp('frisk-backend')

  // Access policies - outbound
  + application.withOutboundHttp('api.airtable.com')
  + application.withOutboundHttp('graph.microsoft.com')

  // External secrets for environment variables
  + application.withEnvironmentVariablesFromExternalSecret(
    esoName,
    secrets=[
      { fromSecret: 'demo-airtable-token', toKey: 'AIRTABLE_ACCESS_TOKEN' },
      { fromSecret: 'demo-airtable-token', toKey: 'DE_SCHEMA_DRIFTSKONTINUITET_AIRTABLE_ACCESS_TOKEN' },
      { fromSecret: 'demo-superuser', toKey: 'DE_OAUTH_SUPER_USER_GROUP' },
      { fromSecret: 'demo-airtable-token', toKey: 'DE_SCHEMA_SIKKERHETSKONTROLLER_AIRTABLE_ACCESS_TOKEN' },
      { fromSecret: 'demo-airtable-token', toKey: 'DE_AIRTABLE_ACCESS_TOKEN' },
    ],
    secretStoreRef=secretStoreName
  )


  // Azure AD Application
  + application.withAzureAdApplication(
    name='demo-service-entraid',
    namespace='demo-main',
    groups=[{ id: 'adsfdsa-gdadf-12346g-dadhjj' }],
    secretPrefix='DE',
    replyUrls=[replyUrl],
    preAuthorizedApplications=[
      {
        cluster: ' ',
        namespace: ' ',
        application: if env == 'dev' then '123456789-asdfghjkl-ddsddw'
        else if env == 'prod' then 'asdfg-12345-qwerty-45678',
      },
    ]
  )

  // config som ikke er støttet av argokit, kan skrives som 'vanlig' jsonnet
  + {
    application+: {
      spec+: {
        gcp: cloudSqlConfig,
        filesFrom: [
          {
            mountPath: '/etc/demo/provisioning/schemasources',
            secret: 'provisioning-file',
          },
        ],
      },
    },
    objects+:: [
      argokit.externalSecrets.store.new(secretStoreName, gsmProjectId),

      argokit.externalSecrets.secret.new(
        'provisioning-file',
        secrets=[
          { fromSecret: 'demo-defaults', toKey: 'defaults.yaml' },
        ],
        secretStoreRef=secretStoreName
      ),

      {
        apiVersion: 'networking.istio.io/v1',
        kind: 'DestinationRule',
        metadata: {
          name: 'istio-sticky' + name,
        },
        spec: {
          host: name,
          trafficPolicy: {
            loadBalancer: {
              consistentHash: {
                httpCookie: {
                  name: 'ISTIO-STICKY',
                  path: '/',
                  ttl: '0',
                },
              },
            },
          },
        },
      },
    ],
  }

## Renders to this Kubernetes Application manifest:

  {
   "apiVersion": "v1",
   "items": [
      {
         "apiVersion": "skiperator.kartverket.no/v1alpha1",
         "kind": "Application",
         "metadata": {
            "name": "demo-backend"
         },
         "spec": {
            "accessPolicy": {
               "inbound": {
                  "rules": [
                     {
                        "application": "demo-frontend"
                     },
                     {
                        "application": "frisk-backend"
                     }
                  ]
               },
               "outbound": {
                  "external": [
                     {
                        "host": "api.airtable.com"
                     },
                     {
                        "host": "graph.microsoft.com"
                     },
                     {
                        "host": "login.microsoftonline.com"
                     }
                  ]
               }
            },
            "env": [
               {
                  "name": "AUTH_PROVIDER_URL",
                  "value": "https://demo.atgcp1-dev.host.com/callback"
               },
               {
                  "name": "DB_NAME",
                  "value": "demo_user"
               },
               {
                  "name": "DB_PASSWORD",
                  "value": ""
               },
               {
                  "name": "FRISK_FRONTEND_URL_HOST",
                  "value": "https://frisk.atgcp1-dev.host.com"
               },
               {
                  "name": "FRONTEND_URL_HOST",
                  "value": "demo.atgcp1-dev.host.com"
               },
               {
                  "name": "RR_DATABASE_HOST",
                  "value": "localhost:5432"
               },
               {
                  "name": "RR_DATABASE_NAME",
                  "value": "demo"
               },
               {
                  "name": "RR_DATABASE_PASSWORD",
                  "value": ""
               },
               {
                  "name": "RR_DATABASE_USER",
                  "value": "demo_user"
               },
               {
                  "name": "RR_OAUTH_TENANT_ID",
                  "value": "f9f9f-abcd-1234-gagaga-lgtm-123"
               },
               {
                  "name": "RR_SERVER_ALLOWED_ORIGINS",
                  "value": "frisk.atgcp1-dev.host.com"
               },
               {
                  "name": "RR_SERVER_DOMAIN",
                  "value": "demo.atgcp1-dev.host.com"
               },
               {
                  "name": "RR_SERVER_HTTP_PORT",
                  "value": "8080"
               },
               {
                  "name": "RR_SERVER_PROTOCOL",
                  "value": "https"
               },
               {
                  "name": "RR_SERVER_ROOT_URL",
                  "value": "https://demo.atgcp1-dev.host.com"
               },
               {
                  "name": "TENANT_ID",
                  "value": "ogaboga-abcd-1234-gagaga-lgtm-123"
               },
               {
                  "name": "RR_OAUTH_CLIENT_ID",
                  "valueFrom": {
                     "secretKeyRef": {
                        "key": "RR_APP_CLIENT_ID",
                        "name": "entraid-secret"
                     }
                  }
               },
               {
                  "name": "RR_OAUTH_CLIENT_SECRET",
                  "valueFrom": {
                     "secretKeyRef": {
                        "key": "RR_APP_CLIENT_SECRET",
                        "name": "entraid-secret"
                     }
                  }
               }
            ],
            "envFrom": [
               {
                  "secret": "devex-demo-secrets"
               },
               {
                  "secret": "DE-demo-service-entraid"
               }
            ],
            "filesFrom": [
               {
                  "mountPath": "/etc/demo/provisioning/schemasources",
                  "secret": "provisioning-file"
               }
            ],
            "gcp": {
               "cloudSqlProxyImage": "gcr.io/cloudsql-docker/gce-proxy:1.33.3",
               "databaseName": "demo_db",
               "databaseUser": "demo_user",
               "instanceConnectionName": "project:region:instance"
            },
            "image": "1",
            "ingresses": [
               "backend.atgcp1-devhost.com"
            ],
            "liveness": {
               "failureThreshold": 3,
               "initialDelay": 0,
               "path": "/health",
               "port": 8080,
               "timeout": 1
            },
            "port": 8080,
            "readiness": {
               "failureThreshold": 3,
               "initialDelay": 0,
               "path": "/health",
               "port": 8080,
               "timeout": 1
            }
         }
      },
      {
         "apiVersion": "nais.io/v1",
         "kind": "AzureAdApplication",
         "metadata": {
            "name": "demo-service-entraid",
            "namespace": "demo-main"
         },
         "spec": {
            "allowAllUsers": false,
            "claims": {
               "groups": [
                  {
                     "id": "adsfdsa-gdadf-12346g-dadhjj"
                  }
               ]
            },
            "preAuthorizedApplications": [
               {
                  "application": "123456789-asdfghjkl-ddsddw",
                  "cluster": " ",
                  "namespace": " "
               }
            ],
            "replyUrls": [
               {
                  "url": "https://demo.atgcp1-dev.host.com/callback"
               },
               {
                  "url": "http://localhost/callback"
               }
            ],
            "secretName": "DE-demo-service-entraid"
         }
      },
      {
         "apiVersion": "external-secrets.io/v1",
         "kind": "SecretStore",
         "metadata": {
            "name": "devex-demo-gsm"
         },
         "spec": {
            "provider": {
               "gcpsm": {
                  "projectID": "very-cool-version"
               }
            }
         }
      },
      {
         "apiVersion": "external-secrets.io/v1",
         "kind": "ExternalSecret",
         "metadata": {
            "name": "devex-demo-secrets"
         },
         "spec": {
            "data": [
               {
                  "remoteRef": {
                     "key": "demo-airtable-token",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "AIRTABLE_ACCESS_TOKEN"
               },
               {
                  "remoteRef": {
                     "key": "demo-airtable-token",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "DE_SCHEMA_DRIFTSKONTINUITET_AIRTABLE_ACCESS_TOKEN"
               },
               {
                  "remoteRef": {
                     "key": "demo-superuser",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "DE_OAUTH_SUPER_USER_GROUP"
               },
               {
                  "remoteRef": {
                     "key": "demo-airtable-token",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "DE_SCHEMA_SIKKERHETSKONTROLLER_AIRTABLE_ACCESS_TOKEN"
               },
               {
                  "remoteRef": {
                     "key": "demo-airtable-token",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "DE_AIRTABLE_ACCESS_TOKEN"
               }
            ],
            "refreshInterval": "1h",
            "secretStoreRef": {
               "kind": "SecretStore",
               "name": "devex-demo-gsm"
            },
            "target": {
               "name": "devex-demo-secrets"
            }
         }
      },
      {
         "apiVersion": "networking.istio.io/v1",
         "kind": "DestinationRule",
         "metadata": {
            "name": "istio-stickydemo-backend"
         },
         "spec": {
            "host": "demo-backend",
            "trafficPolicy": {
               "loadBalancer": {
                  "consistentHash": {
                     "httpCookie": {
                        "name": "ISTIO-STICKY",
                        "path": "/",
                        "ttl": "0"
                     }
                  }
               }
            }
         }
      },
      {
         "apiVersion": "external-secrets.io/v1",
         "kind": "ExternalSecret",
         "metadata": {
            "name": "provisioning-file"
         },
         "spec": {
            "data": [
               {
                  "remoteRef": {
                     "key": "demo-defaults",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "defaults.yaml"
               }
            ],
            "refreshInterval": "1h",
            "secretStoreRef": {
               "kind": "SecretStore",
               "name": "devex-demo-gsm"
            },
            "target": {
               "name": "provisioning-file"
            }
         }
      }
   ],
   "kind": "List"
}