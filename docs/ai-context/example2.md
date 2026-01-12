The files:

# regelrett.jsonnet
local application = import '../../../applications/regelrett.libsonnet';
local cloudSqlConfig = import '../../../utils/cloudSql.libsonnet';
local gcpAuth = import '../../../utils/gcpAuth.libsonnet';
local version = import 'image-url-regelrett';

application(
  env='dev',
  version=version,
  gsmProjectId='skvis-dev-5aa6',
  cloudSqlConfig=cloudSqlConfig(
    connectionName='skvis-dev-5aa6:europe-north1:skvis-pg-01-dev',
    ip='10.144.16.3',
    serviceAccount='skvis-pg-01-regelrett@skvis-dev-5aa6.iam.gserviceaccount.com',
  ),
  dbUser='skvis-pg-01-regelrett@skvis-dev-5aa6.iam',
  replyUrl='https://regelrett.atgcp1-dev.kartverket-intern.cloud/callback'
)

# regelrett.libsonnet
local argokit = import '../argokit/v2/jsonnet/argokit.libsonnet';
local application = argokit.appAndObjects.application;

local cloudSqlConfig = import '../utils/cloudSql.libsonnet';
local gcpAuth = import '../utils/gcpAuth.libsonnet';
local version = import 'image-url-regelrett';

function(
  env,
  name='regelrett-backend',
  version,
  gsmProjectId,
  secretStoreName='regelrett-backend-skvis-gsm',
  esoName='regelrett-backend-secrets',
  kubernetesSecretEntraIdSecretName='entraid-secret-rr',
  cloudSqlConfig,
  dbUser,
  replyUrl,
)
  application.new(name=name, image=version, port=8080)
  + application.withLiveness({path: '/health', port: 8080})
  + application.withReadiness({path: '/health', port: 8080})
  + application.forHostnames('regelrett.atgcp1-' + env + '.kartverket-intern.cloud')
  + application.withEnvironmentVariables({
    RR_SERVER_DOMAIN: 'regelrett.atgcp1-' + env + '.kartverket-intern.cloud',
    RR_SERVER_PROTOCOL: 'https',
    RR_SERVER_HTTP_PORT: '8080',
    RR_SERVER_ROOT_URL: 'https://regelrett.atgcp1-' + env + '.kartverket-intern.cloud',
    FRONTEND_URL_HOST: 'regelrett.atgcp1-' + env + '.kartverket-intern.cloud',
    TENANT_ID: '7f74c8a2-43ce-46b2-b0e8-b6306cba73a3',
    AUTH_PROVIDER_URL: 'https://regelrett.atgcp1-' + env + '.kartverket-intern.cloud/callback',
    DB_NAME: dbUser,
    DB_PASSWORD: '',
    FRISK_FRONTEND_URL_HOST: 'https://frisk.atgcp1-' + env + '.kartverket-intern.cloud',
    RR_OAUTH_TENANT_ID: '7f74c8a2-43ce-46b2-b0e8-b6306cba73a3',
    RR_DATABASE_HOST: 'localhost:5432',
    RR_DATABASE_NAME: 'regelrett',
    RR_DATABASE_USER: dbUser,
    RR_DATABASE_PASSWORD: '',
    RR_SERVER_ALLOWED_ORIGINS: 'frisk.atgcp1-' + env + '.kartverket-intern.cloud',
    RR_MICROSOFT_GRAPH_GROUPFILTER: "startswith(displayName,'AAD -') or displayName eq 'Kartverket' or displayName eq 'ED Eiendomsdivisjonen'",
  })
  + application.withEnvironmentVariableFromSecret('RR_OAUTH_CLIENT_ID', kubernetesSecretEntraIdSecretName, 'RR_APP_CLIENT_ID')
  + application.withEnvironmentVariableFromSecret('RR_OAUTH_CLIENT_SECRET', kubernetesSecretEntraIdSecretName, 'RR_APP_CLIENT_SECRET')
  + application.withInboundSkipApp('regelrett-frontend')
  + application.withInboundSkipApp('frisk-backend')
  + application.withOutboundHttp('api.airtable.com')
  + application.withOutboundHttp('login.microsoftonline.com')
  + application.withOutboundHttp('graph.microsoft.com')
  + {
    application+: {
      spec+: {
        resources: {
          requests: {
            cpu: '25m',
            memory: '256Mi',
          },
        },
        gcp: {
          cloudSqlProxy: cloudSqlConfig,
        },
        filesFrom: [
          {
            mountPath: '/etc/regelrett/provisioning/schemasources',
            secret: 'provisioning-file',
          },
        ],
        envFrom: [
          {secret: esoName},
          {secret: kubernetesSecretEntraIdSecretName},
        ],
      },
    },
    objects+: [
      {
        apiVersion: 'nais.io/v1',
        kind: 'AzureAdApplication',
        metadata: {
          name: 'regelrett-service-entraid',
          namespace: 'regelrett-main',
        },
        spec: {
          secretName: kubernetesSecretEntraIdSecretName,
          secretKeyPrefix: 'RR_',
          claims: {
            groups: [{id: 'a6578085-e0ee-4168-bf97-1b3026f6f3bd'}],
          },
          replyUrls: [
            {url: replyUrl},
          ],
        },
      },
      argokit.externalSecrets.secret.new(
        name=esoName,
        secrets=[
          {fromSecret: 'regelrett-airtable-token', toKey: 'AIRTABLE_ACCESS_TOKEN'},
          {fromSecret: 'regelrett-airtable-token', toKey: 'RR_SCHEMA_DRIFTSKONTINUITET_AIRTABLE_ACCESS_TOKEN'},
          {fromSecret: 'regelrett-superuser', toKey: 'RR_OAUTH_SUPER_USER_GROUP'},
          {fromSecret: 'regelrett-airtable-token', toKey: 'RR_SCHEMA_SIKKERHETSKONTROLLER_AIRTABLE_ACCESS_TOKEN'},
          {fromSecret: 'regelrett-airtable-token', toKey: 'RR_AIRTABLE_ACCESS_TOKEN'},
        ],
        secretStoreRef=secretStoreName
      ),
      argokit.externalSecrets.secret.new(
        name='provisioning-file',
        secrets=[
          {fromSecret: 'regelrett-defaults', toKey: 'defaults.yaml'},
        ],
        secretStoreRef=secretStoreName
      ),
      argokit.externalSecrets.store.new(
        name=secretStoreName,
        gcpProject=gsmProjectId
      ),
      {
        apiVersion: 'networking.istio.io/v1',
        kind: 'DestinationRule',
        metadata: {name: 'istio-sticky' + name},
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

Renders to:
{
   "apiVersion": "v1",
   "items": [
      {
         "apiVersion": "networking.istio.io/v1",
         "kind": "DestinationRule",
         "metadata": {
            "name": "istio-stickyregelrett-backend"
         },
         "spec": {
            "host": "regelrett-backend",
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
            "labels": {
               "skip.kartverket.no/argokit-flavor": "v2",
               "skip.kartverket.no/argokit-git-ref": "ed4782236a67d46420f51434de68c26d478fc241",
               "skip.kartverket.no/argokit-tag": "dev-dirty"
            },
            "name": "provisioning-file"
         },
         "spec": {
            "data": [
               {
                  "remoteRef": {
                     "key": "regelrett-defaults",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "defaults.yaml"
               }
            ],
            "refreshInterval": "1h0m0s",
            "secretStoreRef": {
               "kind": "SecretStore",
               "name": "regelrett-backend-skvis-gsm"
            },
            "target": {
               "name": "provisioning-file"
            }
         }
      },
      {
         "apiVersion": "skiperator.kartverket.no/v1alpha1",
         "kind": "Application",
         "metadata": {
            "labels": {
               "skip.kartverket.no/argokit-flavor": "v2",
               "skip.kartverket.no/argokit-git-ref": "ed4782236a67d46420f51434de68c26d478fc241",
               "skip.kartverket.no/argokit-tag": "dev-dirty"
            },
            "name": "regelrett-backend"
         },
         "spec": {
            "accessPolicy": {
               "inbound": {
                  "rules": [
                     {
                        "application": "regelrett-frontend"
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
                        "host": "login.microsoftonline.com"
                     },
                     {
                        "host": "graph.microsoft.com"
                     }
                  ]
               }
            },
            "env": [
               {
                  "name": "AUTH_PROVIDER_URL",
                  "value": "https://regelrett.atgcp1-dev.kartverket-intern.cloud/callback"
               },
               {
                  "name": "DB_NAME",
                  "value": "skvis-pg-01-regelrett@skvis-dev-5aa6.iam"
               },
               {
                  "name": "DB_PASSWORD",
                  "value": ""
               },
               {
                  "name": "FRISK_FRONTEND_URL_HOST",
                  "value": "https://frisk.atgcp1-dev.kartverket-intern.cloud"
               },
               {
                  "name": "FRONTEND_URL_HOST",
                  "value": "regelrett.atgcp1-dev.kartverket-intern.cloud"
               },
               {
                  "name": "RR_DATABASE_HOST",
                  "value": "localhost:5432"
               },
               {
                  "name": "RR_DATABASE_NAME",
                  "value": "regelrett"
               },
               {
                  "name": "RR_DATABASE_PASSWORD",
                  "value": ""
               },
               {
                  "name": "RR_DATABASE_USER",
                  "value": "skvis-pg-01-regelrett@skvis-dev-5aa6.iam"
               },
               {
                  "name": "RR_MICROSOFT_GRAPH_GROUPFILTER",
                  "value": "startswith(displayName,'AAD -') or displayName eq 'Kartverket' or displayName eq 'ED Eiendomsdivisjonen'"
               },
               {
                  "name": "RR_OAUTH_TENANT_ID",
                  "value": "7f74c8a2-43ce-46b2-b0e8-b6306cba73a3"
               },
               {
                  "name": "RR_SERVER_ALLOWED_ORIGINS",
                  "value": "frisk.atgcp1-dev.kartverket-intern.cloud"
               },
               {
                  "name": "RR_SERVER_DOMAIN",
                  "value": "regelrett.atgcp1-dev.kartverket-intern.cloud"
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
                  "value": "https://regelrett.atgcp1-dev.kartverket-intern.cloud"
               },
               {
                  "name": "TENANT_ID",
                  "value": "7f74c8a2-43ce-46b2-b0e8-b6306cba73a3"
               },
               {
                  "name": "RR_OAUTH_CLIENT_ID",
                  "valueFrom": {
                     "secretKeyRef": {
                        "key": "RR_APP_CLIENT_ID",
                        "name": "entraid-secret-rr"
                     }
                  }
               },
               {
                  "name": "RR_OAUTH_CLIENT_SECRET",
                  "valueFrom": {
                     "secretKeyRef": {
                        "key": "RR_APP_CLIENT_SECRET",
                        "name": "entraid-secret-rr"
                     }
                  }
               }
            ],
            "envFrom": [
               {
                  "secret": "regelrett-backend-secrets"
               },
               {
                  "secret": "entraid-secret-rr"
               }
            ],
            "filesFrom": [
               {
                  "mountPath": "/etc/regelrett/provisioning/schemasources",
                  "secret": "provisioning-file"
               }
            ],
            "gcp": {
               "cloudSqlProxy": {
                  "connectionName": "skvis-dev-5aa6:europe-north1:skvis-pg-01-dev",
                  "ip": "10.144.16.3",
                  "serviceAccount": "skvis-pg-01-regelrett@skvis-dev-5aa6.iam.gserviceaccount.com"
               }
            },
            "image": "ghcr.io/kartverket/regelrett:main",
            "ingresses": [
               "regelrett.atgcp1-dev.kartverket-intern.cloud"
            ],
            "liveness": {
               "path": "/health",
               "port": 8080
            },
            "port": 8080,
            "readiness": {
               "path": "/health",
               "port": 8080
            },
            "resources": {
               "requests": {
                  "cpu": "25m",
                  "memory": "256Mi"
               }
            }
         }
      },
      {
         "apiVersion": "external-secrets.io/v1",
         "kind": "ExternalSecret",
         "metadata": {
            "labels": {
               "skip.kartverket.no/argokit-flavor": "v2",
               "skip.kartverket.no/argokit-git-ref": "ed4782236a67d46420f51434de68c26d478fc241",
               "skip.kartverket.no/argokit-tag": "dev-dirty"
            },
            "name": "regelrett-backend-secrets"
         },
         "spec": {
            "data": [
               {
                  "remoteRef": {
                     "key": "regelrett-airtable-token",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "AIRTABLE_ACCESS_TOKEN"
               },
               {
                  "remoteRef": {
                     "key": "regelrett-airtable-token",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "RR_SCHEMA_DRIFTSKONTINUITET_AIRTABLE_ACCESS_TOKEN"
               },
               {
                  "remoteRef": {
                     "key": "regelrett-superuser",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "RR_OAUTH_SUPER_USER_GROUP"
               },
               {
                  "remoteRef": {
                     "key": "regelrett-airtable-token",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "RR_SCHEMA_SIKKERHETSKONTROLLER_AIRTABLE_ACCESS_TOKEN"
               },
               {
                  "remoteRef": {
                     "key": "regelrett-airtable-token",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "RR_AIRTABLE_ACCESS_TOKEN"
               }
            ],
            "refreshInterval": "1h0m0s",
            "secretStoreRef": {
               "kind": "SecretStore",
               "name": "regelrett-backend-skvis-gsm"
            },
            "target": {
               "name": "regelrett-backend-secrets"
            }
         }
      },
      {
         "apiVersion": "external-secrets.io/v1",
         "kind": "SecretStore",
         "metadata": {
            "name": "regelrett-backend-skvis-gsm"
         },
         "spec": {
            "provider": {
               "gcpsm": {
                  "projectID": "skvis-dev-5aa6"
               }
            }
         }
      },
      {
         "apiVersion": "nais.io/v1",
         "kind": "AzureAdApplication",
         "metadata": {
            "name": "regelrett-service-entraid",
            "namespace": "regelrett-main"
         },
         "spec": {
            "claims": {
               "groups": [
                  {
                     "id": "a6578085-e0ee-4168-bf97-1b3026f6f3bd"
                  }
               ]
            },
            "replyUrls": [
               {
                  "url": "https://regelrett.atgcp1-dev.kartverket-intern.cloud/callback"
               }
            ],
            "secretKeyPrefix": "RR_",
            "secretName": "entraid-secret-rr"
         }
      }
   ],
   "kind": "List"
}