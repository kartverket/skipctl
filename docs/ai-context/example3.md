The files:

# frisk-backend.jsonnet
local application = import '../../../applications-argokitv2/frisk-backend.libsonnet';
local version = import 'image-url-frisk-backend';

application(
    env='dev',
    version=version,
    clientId='3129a75a-aad3-4fec-b972-61d1b7c21e6c',
    gsmProjectId='skvis-dev-5aa6',
    databaseHost='10.144.16.15',
)

# frisk-backend.libsonnet
local argokit = import '../argokit/v2/jsonnet/argokit.libsonnet';
local application = argokit.appAndObjects.application;
local esoName = 'frisk-backend-secrets';

function(env, version, name='frisk-backend', clientId, secretStoreName='frisk-backend-skvis-gsm', esoName='frisk-backend-secrets', gsmProjectId, databaseHost)
  local secretEsoName =
  argokit.externalSecrets.secret.new(
        name=esoName,
        secrets=[
          {fromSecret: 'frisk-client-secret-backend', toKey: 'CLIENT_SECRET', conversionStrategy:null, decodingStrategy:null},
          {fromSecret: 'cloudsql-frisk-backend-db-private-ip', toKey: 'DATABASE_HOST', conversionStrategy:null, decodingStrategy:null},
          {fromSecret: 'cloudsql-frisk-backend-jdbc-url', toKey: 'JDBC_URL', conversionStrategy:null, decodingStrategy:null},
          {fromSecret: 'cloudsql-frisk-backend-db-admin-password', toKey: 'DATABASE_PASSWORD', conversionStrategy:null, decodingStrategy:null},
        ],
        secretStoreRef=secretStoreName
      );
  local externalSecretStore =
    argokit.externalSecrets.store.new(
      name=secretStoreName,
      gcpProject=gsmProjectId
    );
  local networkIstio = {
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
      };
    local externalSecret = {
        apiVersion: 'external-secrets.io/v1',
        kind: 'ExternalSecret',
        metadata: {
          name: 'db-ssl-ca',
        },
        spec: {
          data: [
            {
              remoteRef: {
                key: 'cloudsql-frisk-backend-db-ca-certificate',
                metadataPolicy: 'None',
              },
              secretKey: 'server-ca.pem',
            },
            {
              remoteRef: {
                key: 'cloudsql-frisk-backend-db-admin-client-certificate',
                metadataPolicy: 'None',
              },
              secretKey: 'client-cert.pem',
            },
            {
              remoteRef: {
                key: 'cloudsql-frisk-backend-db-admin-client-key',
                metadataPolicy: 'None',
              },
              secretKey: 'client-key.key',
            },
            {
              remoteRef: {
                key: 'cloudsql-frisk-backend-db-admin-client-key-pk8',
                metadataPolicy: 'None',
              },
              secretKey: 'client-key.pk8',
            },
          ],
          refreshInterval: '1h',
          secretStoreRef: {
            kind: 'SecretStore',
            name: secretStoreName,
          },
          target: {
            name: 'db-ssl-ca',
          },
        },
      };
    local networking = {
        apiVersion: 'networking.k8s.io/v1',
        kind: 'NetworkPolicy',
        metadata: {
          name: 'frisk-backend-cloudsql',
          annotations: {
            'argocd.argoproj.io/sync-options': 'Prune=false',
          },
        },
        spec: {
          egress: [
            {
              ports: [
                {
                  port: 5432,
                  protocol: 'TCP',
                },
              ],
              to: [
                {
                  ipBlock: {
                    cidr: databaseHost + '/32',
                  },
                },
              ],
            },
          ],
          podSelector: {
            matchExpressions: [
              {
                key: 'app',
                operator: 'In',
                values: [name],
              },
            ],
          },
          policyTypes: ['Egress'],
        },
      };
  
  application.new(name=name, image=version, port=8080)
  + application.withLiveness(application.probe(path='/health', port=8080))
  + application.withReadiness(application.probe(path='/health', port=8080))
  + application.forHostnames('api.frisk.atgcp1-' + env + '.kartverket-intern.cloud')
  + application.resources.withRequests(cpu = '25m', memory = '256Mi')
  + application.withEnvironmentVariables(
    {
      clientId: clientId,
      tenantId: '7f74c8a2-43ce-46b2-b0e8-b6306cba73a3',
      environment: 'production',
      skipEnv: env,
      DATABASE_USER: 'postgres',
      REGELRETT_URL: 'http://regelrett-backend.regelrett-main:8080',
      DATABASE_USERNAME: 'admin',
      ALLOWED_CORS_HOSTS: 'frisk.atgcp1-' + env + '.kartverket-intern.cloud',
    }
  )
  + application.withInboundSkipApp(appname='frisk-frontend')
  + application.withOutboundSkipApp(appname='regelrett-backend')
  + application.withOutboundPostgres(host='frisk-backend-db-dev', ip='10.144.16.15')
  + application.withOutboundHttp(host='login.microsoftonline.com')
  + application.withOutboundHttp(host='graph.microsoft.com')
  + application.withEnvironmentVariablesFromSecret(secretName=esoName)
  + application.withObjects([secretEsoName, externalSecretStore, networkIstio, externalSecret, networking])
  + {
    application+: {
      spec+: {
        filesFrom: [
          {
            mountPath: '/app/db-ssl-ca',
            secret: 'db-ssl-ca',
          },
        ],
      },
    },
  }


Renders to:
{
   "apiVersion": "v1",
   "items": [
      {
         "apiVersion": "external-secrets.io/v1",
         "kind": "ExternalSecret",
         "metadata": {
            "name": "db-ssl-ca"
         },
         "spec": {
            "data": [
               {
                  "remoteRef": {
                     "key": "cloudsql-frisk-backend-db-ca-certificate",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "server-ca.pem"
               },
               {
                  "remoteRef": {
                     "key": "cloudsql-frisk-backend-db-admin-client-certificate",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "client-cert.pem"
               },
               {
                  "remoteRef": {
                     "key": "cloudsql-frisk-backend-db-admin-client-key",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "client-key.key"
               },
               {
                  "remoteRef": {
                     "key": "cloudsql-frisk-backend-db-admin-client-key-pk8",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "client-key.pk8"
               }
            ],
            "refreshInterval": "1h",
            "secretStoreRef": {
               "kind": "SecretStore",
               "name": "frisk-backend-skvis-gsm"
            },
            "target": {
               "name": "db-ssl-ca"
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
            "name": "frisk-backend"
         },
         "spec": {
            "accessPolicy": {
               "inbound": {
                  "rules": [
                     {
                        "application": "frisk-frontend"
                     }
                  ]
               },
               "outbound": {
                  "external": [
                     {
                        "host": "frisk-backend-db-dev",
                        "ip": "10.144.16.15",
                        "ports": [
                           {
                              "name": "postgres-port",
                              "port": 5432,
                              "protocol": "TCP"
                           }
                        ]
                     },
                     {
                        "host": "login.microsoftonline.com"
                     },
                     {
                        "host": "graph.microsoft.com"
                     }
                  ],
                  "rules": [
                     {
                        "application": "regelrett-backend"
                     }
                  ]
               }
            },
            "env": [
               {
                  "name": "ALLOWED_CORS_HOSTS",
                  "value": "frisk.atgcp1-dev.kartverket-intern.cloud"
               },
               {
                  "name": "DATABASE_USER",
                  "value": "postgres"
               },
               {
                  "name": "DATABASE_USERNAME",
                  "value": "admin"
               },
               {
                  "name": "REGELRETT_URL",
                  "value": "http://regelrett-backend.regelrett-main:8080"
               },
               {
                  "name": "clientId",
                  "value": "3129a75a-aad3-4fec-b972-61d1b7c21e6c"
               },
               {
                  "name": "environment",
                  "value": "production"
               },
               {
                  "name": "skipEnv",
                  "value": "dev"
               },
               {
                  "name": "tenantId",
                  "value": "7f74c8a2-43ce-46b2-b0e8-b6306cba73a3"
               }
            ],
            "envFrom": [
               {
                  "secret": "frisk-backend-secrets"
               }
            ],
            "filesFrom": [
               {
                  "mountPath": "/app/db-ssl-ca",
                  "secret": "db-ssl-ca"
               }
            ],
            "image": "ghcr.io/kartverket/frisk-backend@sha256:b377d6306b304e44141d454acc8418ac5e55cf56ed2cae97e002a95b49c0a28b",
            "ingresses": [
               "api.frisk.atgcp1-dev.kartverket-intern.cloud"
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
         "apiVersion": "networking.k8s.io/v1",
         "kind": "NetworkPolicy",
         "metadata": {
            "annotations": {
               "argocd.argoproj.io/sync-options": "Prune=false"
            },
            "name": "frisk-backend-cloudsql"
         },
         "spec": {
            "egress": [
               {
                  "ports": [
                     {
                        "port": 5432,
                        "protocol": "TCP"
                     }
                  ],
                  "to": [
                     {
                        "ipBlock": {
                           "cidr": "10.144.16.15/32"
                        }
                     }
                  ]
               }
            ],
            "podSelector": {
               "matchExpressions": [
                  {
                     "key": "app",
                     "operator": "In",
                     "values": [
                        "frisk-backend"
                     ]
                  }
               ]
            },
            "policyTypes": [
               "Egress"
            ]
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
            "name": "frisk-backend-secrets"
         },
         "spec": {
            "data": [
               {
                  "remoteRef": {
                     "key": "frisk-client-secret-backend",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "CLIENT_SECRET"
               },
               {
                  "remoteRef": {
                     "key": "cloudsql-frisk-backend-db-private-ip",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "DATABASE_HOST"
               },
               {
                  "remoteRef": {
                     "key": "cloudsql-frisk-backend-jdbc-url",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "JDBC_URL"
               },
               {
                  "remoteRef": {
                     "key": "cloudsql-frisk-backend-db-admin-password",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "DATABASE_PASSWORD"
               }
            ],
            "refreshInterval": "1h0m0s",
            "secretStoreRef": {
               "kind": "SecretStore",
               "name": "frisk-backend-skvis-gsm"
            },
            "target": {
               "name": "frisk-backend-secrets"
            }
         }
      },
      {
         "apiVersion": "external-secrets.io/v1",
         "kind": "SecretStore",
         "metadata": {
            "name": "frisk-backend-skvis-gsm"
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
         "apiVersion": "networking.istio.io/v1",
         "kind": "DestinationRule",
         "metadata": {
            "name": "istio-stickyfrisk-backend"
         },
         "spec": {
            "host": "frisk-backend",
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
      }
   ],
   "kind": "List"
}