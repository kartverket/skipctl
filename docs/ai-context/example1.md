These files: 

# sikkerhetsmetrikker.jsonnet

local application = import '../../../applications/sikkerhetsmetrikker.libsonnet';
local gcpAuth = import '../../../utils/gcpAuth.libsonnet';
local version = import 'image-url-sikkerhetsmetrikker';

application(
  env='dev',
  name='sikkerhetsmetrikker',
  version=version,
  gcpAuth=gcpAuth(
    serviceAccount='sikkerhetsmetrikker@skvis-dev-5aa6.iam.gserviceaccount.com',
  ),
  clientId='efdaa164-4b95-4a40-b3d3-de4e552524e7',
  gsmProjectId='skvis-dev-5aa6',
  dbHost='10.144.16.251',
)

# sikkerhetsmetrikker.libsonnet
local argokit = import '../argokit/v2/jsonnet/argokit.libsonnet';
local application = argokit.appAndObjects.application;

local image = import 'image-url-sikkerhetsmetrikker';
local gcpAuth = import '../../../utils/gcpAuth.libsonnet';

function(
  env,
  name,
  version,
  gcpAuth,
  clientId,
  gsmProjectId,
  gsm='smapi-gsm',
  dbHost
)
  application.new(name=name, image=version, port=8080)
  // Liveness and Readiness probes from source manifest
  + application.withLiveness(application.probe(path='/actuator/health', port=8081, failureThreshold=10, initialDelay=30))
  + application.withReadiness(application.probe(path='/actuator/health', port=8081, failureThreshold=10, initialDelay=30))

  // Environment variables from source manifest
  + application.withEnvironmentVariables({
    GSM_PROJECT_ID: gsmProjectId,
    DB_USERNAME: 'smapiapplication',
  })

  // Environment variables from secret
  + application.withEnvironmentVariablesFromSecret('smapi-db-user-password')

  // Ingress from source manifest
  + application.forHostnames('sikkerhetsmetrikker.atgcp1-' + env + '.kartverket-intern.cloud')

  // Access policies - inbound from source manifest
  + application.withInboundSkipApp('backstage', 'backstage')
  + application.withInboundSkipApp('nightly-update-skipjob')

  // Access policies - outbound from source manifest
  + application.withOutboundHttp('api.github.com')
  + application.withOutboundHttp('eu1.app.sysdig.com')
  + application.withOutboundHttp('kartverket.dev')
  + application.withOutboundHttp('login.microsoftonline.com')
  + application.withOutboundHttp('hooks.slack.com')
  + application.withOutboundHttp('slack.com')
  + application.withOutboundHttp('cloud.tenable.com')
  + application.withOutboundPostgres(host=if env == 'dev' then 'smapi-db-dev' else if env == 'prod' then 'smapi-db-prod', ip=dbHost)
  + application.withOutboundSkipApp('security-champion-api')

  // Additional Kubernetes objects
  + {
    application+: {
      spec+: {
        accessPolicy+: {
          outbound+: {
            // TODO: Spørr. Er dette nødvendig?? Mapper for å endre på port-navnet fra postgres-port til sql som originalen har
            external: std.map(
              function(ext) 
                if std.objectHas(ext, 'ports') && std.length(ext.ports) > 0 && ext.ports[0].port == 5432 
                then ext + { ports: [{ name: 'sql', port: 5432, protocol: 'TCP' }] }
                else ext,
              super.external
            ),
          },
        },
        prometheus: {
          path: '/actuator/prometheus',
          port: 8081,
        },
        filesFrom: [
          {
            mountPath: '/app/db-ssl-ca',
            secret: 'db-ssl-values',
          },
        ],
        gcp: {
          auth: gcpAuth,
        },
        resources: {
          requests: {
            cpu: '25m',
            memory: '512Mi',
          },
        },
      },
    },
    objects+: [
      // Istio DestinationRule
      {
        apiVersion: 'networking.istio.io/v1',
        kind: 'DestinationRule',
        metadata: {
          name: 'istio-sticky',
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
      // Istio RequestAuthentication
      {
        apiVersion: 'security.istio.io/v1',
        kind: 'RequestAuthentication',
        metadata: {
          name: name + '-auth-n',
        },
        spec: {
          selector: {
            matchLabels: {
              app: name,
            },
          },
          jwtRules: [
            {
              issuer: 'https://login.microsoftonline.com/7f74c8a2-43ce-46b2-b0e8-b6306cba73a3/v2.0',
              jwksUri: 'https://login.microsoftonline.com/7f74c8a2-43ce-46b2-b0e8-b6306cba73a3/discovery/v2.0/keys',
              audiences: [clientId],
              forwardOriginalToken: true,
            },
          ],
        },
      },
      // ExternalSecret for db-ssl-values
      argokit.externalSecrets.secret.new(
        name='db-ssl-values',
        secrets=[
          { fromSecret: 'cloudsql-smapi-db-smapiapplication', toKey: 'client-key.pk8' },
          { fromSecret: 'cloudsql-smapi-db-smapiapplication', toKey: 'client-cert.pem' },
          { fromSecret: 'cloudsql-smapi-db-instance', toKey: 'server-ca.pem' },
        ],
        secretStoreRef=gsm
      ),
      // ExternalSecret for smapi-db-user-password
      argokit.externalSecrets.secret.new(
        name='smapi-db-user-password',
        secrets=[
          { fromSecret: 'cloudsql-smapi-db-smapiapplication', toKey: 'DB_PASSWORD' },
          { fromSecret: 'cloudsql-smapi-db-instance', toKey: 'DB_HOST' },
        ],
        secretStoreRef=gsm
      ),
      // SecretStore
      argokit.externalSecrets.store.new(
        name=gsm,
        gcpProject=gsmProjectId
      ),
      // Istio AuthorizationPolicy
      {
        apiVersion: 'security.istio.io/v1beta1',
        kind: 'AuthorizationPolicy',
        metadata: {
          name: name + '-auth-z',
        },
        spec: {
          selector: {
            matchLabels: {
              app: name,
            },
          },
          action: 'ALLOW',
          rules: [
            {
              to: [
                {
                  operation: {
                    methods: ['GET'],
                    paths: [
                      '/swagger-ui*', '/api-docs*', '/token', '/login/oauth2/code/entra', '/oauth2/authorization/entra', '/dummy/vulnerabilities/*', '/api/dynamicScan/**', '/api/dynamicScan/*', '/api/dynamicScan/',
                    ],
                  },
                },
                {
                  operation: {
                    methods: ['POST'],
                    paths: ['/api/risc/test'],
                  },
                },
              ],
            },
            {
              to: [
                {
                  operation: {
                    methods: ['POST', 'DELETE'],
                    paths: ['/api/oppdateringer*', '/api/securityChampion/update', '/api/slack/slackNotifications'],
                  },
                },
              ],
              when: [
                {
                  key: 'request.auth.claims[roles]',
                  values: ['sikkerhetsmetrikker.skrive.alt'],
                },
              ],
            },
            {
              to: [
                {
                  operation: {
                    methods: ['GET'],
                  },
                },
              ],
              when: [
                {
                  key: 'request.auth.claims[roles]',
                  values: ['sikkerhetsmetrikker.lese.alt'],
                },
              ],
            },
            {
              to: [
                {
                  operation: {
                    methods: ['POST'],
                    paths: ['/api/securityChampion'],
                  },
                },
              ],
              when: [
                {
                  key: 'request.auth.claims[iss]',
                  values: ['https://login.microsoftonline.com/7f74c8a2-43ce-46b2-b0e8-b6306cba73a3/v2.0'],
                },
              ],
            },
            {
              to: [
                {
                  operation: {
                    methods: ['GET'],
                    paths: ['/api/securityChampion/workMail'],
                  },
                },
              ],
              when: [
                {
                  key: 'request.auth.claims[roles]',
                  values: ['sikkerhetsmetrikker.lese.alt', 'githubkvmail.lese.alt'],
                },
              ],
            },
            {
              to: [
                {
                  operation: {
                    methods: ['GET'],
                    paths: ['/api/metrikker/*', '/api/scannerData/*'],
                  },
                },
                {
                  operation: {
                    methods: ['POST'],
                    paths: ['/api/metrikker/vulnerabilities/trends/counts', '/api/securityChampion/workMail', '/api/scannerData', '/api/scannerData/trends', '/api/metrikker/ros-status', '/api/metrikker/ros-status/v2', '/api/dynamicScan/', '/api/dynamicScan/*', '/api/dynamicScan/**', '/api/oppdateringer/alertsMetadata/accept'],
                  },
                },
              ],
              when: [
                {
                  key: 'request.auth.claims[iss]',
                  values: ['https://login.microsoftonline.com/7f74c8a2-43ce-46b2-b0e8-b6306cba73a3/v2.0'],
                },
              ],
            },
            {
              from: [
                {
                  source: {
                    principals: ['cluster.local/ns/sikkerhetsmetrikker-main/sa/nightly-update-skipjob'],
                  },
                },
              ],
              to: [
                {
                  operation: {
                    methods: ['POST', 'GET'],
                    paths: ['/api/oppdateringer*'],
                  },
                },
              ],
            },
            {
              from: [
                {
                  source: {
                    principals: ['cluster.local/ns/sikkerhetsmetrikker-main/sa/nightly-update-skipjob'],
                  },
                },
              ],
              to: [
                {
                  operation: {
                    methods: ['POST'],
                    paths: ['/api/slack/nightlyUpdate', '/api/slack/slackNotifications'],
                  },
                },
              ],
            },
            {
              from: [
                {
                  source: {
                    principals: ['cluster.local/ns/backstage/sa/backstage'],
                  },
                },
              ],
              to: [
                {
                  operation: {
                    methods: ['PUT'],
                    paths: ['/api/slack/configure-notifications'],
                  },
                },
              ],
            },
            {
              to: [
                {
                  operation: {
                    methods: ['POST'],
                    paths: ['/api/backstage/catalogInfo'],
                  },
                },
              ],
            },
            {
              to: [
                {
                  operation: {
                    methods: ['GET'],
                    paths: ['/api/public/metrikker/avdeling'],
                  },
                },
              ],
            },
          ],
        },
      },
    ],
  }


Renders to:

{
   "apiVersion": "v1",
   "items": [
      {
         "apiVersion": "external-secrets.io/v1",
         "kind": "ExternalSecret",
         "metadata": {
            "labels": {
               "skip.kartverket.no/argokit-flavor": "v2",
               "skip.kartverket.no/argokit-git-ref": "ed4782236a67d46420f51434de68c26d478fc241",
               "skip.kartverket.no/argokit-tag": "dev-dirty"
            },
            "name": "db-ssl-values"
         },
         "spec": {
            "data": [
               {
                  "remoteRef": {
                     "key": "cloudsql-smapi-db-smapiapplication",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "client-key.pk8"
               },
               {
                  "remoteRef": {
                     "key": "cloudsql-smapi-db-smapiapplication",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "client-cert.pem"
               },
               {
                  "remoteRef": {
                     "key": "cloudsql-smapi-db-instance",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "server-ca.pem"
               }
            ],
            "refreshInterval": "1h0m0s",
            "secretStoreRef": {
               "kind": "SecretStore",
               "name": "smapi-gsm"
            },
            "target": {
               "name": "db-ssl-values"
            }
         }
      },
      {
         "apiVersion": "networking.istio.io/v1",
         "kind": "DestinationRule",
         "metadata": {
            "name": "istio-sticky"
         },
         "spec": {
            "host": "sikkerhetsmetrikker",
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
         "apiVersion": "skiperator.kartverket.no/v1alpha1",
         "kind": "Application",
         "metadata": {
            "labels": {
               "skip.kartverket.no/argokit-flavor": "v2",
               "skip.kartverket.no/argokit-git-ref": "ed4782236a67d46420f51434de68c26d478fc241",
               "skip.kartverket.no/argokit-tag": "dev-dirty"
            },
            "name": "sikkerhetsmetrikker"
         },
         "spec": {
            "accessPolicy": {
               "inbound": {
                  "rules": [
                     {
                        "application": "backstage",
                        "namespace": "backstage"
                     },
                     {
                        "application": "nightly-update-skipjob"
                     }
                  ]
               },
               "outbound": {
                  "external": [
                     {
                        "host": "api.github.com"
                     },
                     {
                        "host": "eu1.app.sysdig.com"
                     },
                     {
                        "host": "kartverket.dev"
                     },
                     {
                        "host": "login.microsoftonline.com"
                     },
                     {
                        "host": "hooks.slack.com"
                     },
                     {
                        "host": "slack.com"
                     },
                     {
                        "host": "cloud.tenable.com"
                     },
                     {
                        "host": "smapi-db-dev",
                        "ip": "10.144.16.251",
                        "ports": [
                           {
                              "name": "sql",
                              "port": 5432,
                              "protocol": "TCP"
                           }
                        ]
                     }
                  ],
                  "rules": [
                     {
                        "application": "security-champion-api"
                     }
                  ]
               }
            },
            "env": [
               {
                  "name": "DB_USERNAME",
                  "value": "smapiapplication"
               },
               {
                  "name": "GSM_PROJECT_ID",
                  "value": "skvis-dev-5aa6"
               }
            ],
            "envFrom": [
               {
                  "secret": "smapi-db-user-password"
               }
            ],
            "filesFrom": [
               {
                  "mountPath": "/app/db-ssl-ca",
                  "secret": "db-ssl-values"
               }
            ],
            "gcp": {
               "auth": {
                  "serviceAccount": "sikkerhetsmetrikker@skvis-dev-5aa6.iam.gserviceaccount.com"
               }
            },
            "image": "ghcr.io/kartverket/sikkerhetsmetrikker@sha256:bc66d1638a7ce64878499f95f5d1f5d018d32abf3b06e5c3357c219fecb842ca",
            "ingresses": [
               "sikkerhetsmetrikker.atgcp1-dev.kartverket-intern.cloud"
            ],
            "liveness": {
               "failureThreshold": 10,
               "initialDelay": 30,
               "path": "/actuator/health",
               "port": 8081,
               "timeout": 1
            },
            "port": 8080,
            "prometheus": {
               "path": "/actuator/prometheus",
               "port": 8081
            },
            "readiness": {
               "failureThreshold": 10,
               "initialDelay": 30,
               "path": "/actuator/health",
               "port": 8081,
               "timeout": 1
            },
            "resources": {
               "requests": {
                  "cpu": "25m",
                  "memory": "512Mi"
               }
            }
         }
      },
      {
         "apiVersion": "security.istio.io/v1",
         "kind": "RequestAuthentication",
         "metadata": {
            "name": "sikkerhetsmetrikker-auth-n"
         },
         "spec": {
            "jwtRules": [
               {
                  "audiences": [
                     "efdaa164-4b95-4a40-b3d3-de4e552524e7"
                  ],
                  "forwardOriginalToken": true,
                  "issuer": "https://login.microsoftonline.com/7f74c8a2-43ce-46b2-b0e8-b6306cba73a3/v2.0",
                  "jwksUri": "https://login.microsoftonline.com/7f74c8a2-43ce-46b2-b0e8-b6306cba73a3/discovery/v2.0/keys"
               }
            ],
            "selector": {
               "matchLabels": {
                  "app": "sikkerhetsmetrikker"
               }
            }
         }
      },
      {
         "apiVersion": "security.istio.io/v1beta1",
         "kind": "AuthorizationPolicy",
         "metadata": {
            "name": "sikkerhetsmetrikker-auth-z"
         },
         "spec": {
            "action": "ALLOW",
            "rules": [
               {
                  "to": [
                     {
                        "operation": {
                           "methods": [
                              "GET"
                           ],
                           "paths": [
                              "/swagger-ui*",
                              "/api-docs*",
                              "/token",
                              "/login/oauth2/code/entra",
                              "/oauth2/authorization/entra",
                              "/dummy/vulnerabilities/*",
                              "/api/dynamicScan/**",
                              "/api/dynamicScan/*",
                              "/api/dynamicScan/"
                           ]
                        }
                     },
                     {
                        "operation": {
                           "methods": [
                              "POST"
                           ],
                           "paths": [
                              "/api/risc/test"
                           ]
                        }
                     }
                  ]
               },
               {
                  "to": [
                     {
                        "operation": {
                           "methods": [
                              "POST",
                              "DELETE"
                           ],
                           "paths": [
                              "/api/oppdateringer*",
                              "/api/securityChampion/update",
                              "/api/slack/slackNotifications"
                           ]
                        }
                     }
                  ],
                  "when": [
                     {
                        "key": "request.auth.claims[roles]",
                        "values": [
                           "sikkerhetsmetrikker.skrive.alt"
                        ]
                     }
                  ]
               },
               {
                  "to": [
                     {
                        "operation": {
                           "methods": [
                              "GET"
                           ]
                        }
                     }
                  ],
                  "when": [
                     {
                        "key": "request.auth.claims[roles]",
                        "values": [
                           "sikkerhetsmetrikker.lese.alt"
                        ]
                     }
                  ]
               },
               {
                  "to": [
                     {
                        "operation": {
                           "methods": [
                              "POST"
                           ],
                           "paths": [
                              "/api/securityChampion"
                           ]
                        }
                     }
                  ],
                  "when": [
                     {
                        "key": "request.auth.claims[iss]",
                        "values": [
                           "https://login.microsoftonline.com/7f74c8a2-43ce-46b2-b0e8-b6306cba73a3/v2.0"
                        ]
                     }
                  ]
               },
               {
                  "to": [
                     {
                        "operation": {
                           "methods": [
                              "GET"
                           ],
                           "paths": [
                              "/api/securityChampion/workMail"
                           ]
                        }
                     }
                  ],
                  "when": [
                     {
                        "key": "request.auth.claims[roles]",
                        "values": [
                           "sikkerhetsmetrikker.lese.alt",
                           "githubkvmail.lese.alt"
                        ]
                     }
                  ]
               },
               {
                  "to": [
                     {
                        "operation": {
                           "methods": [
                              "GET"
                           ],
                           "paths": [
                              "/api/metrikker/*",
                              "/api/scannerData/*"
                           ]
                        }
                     },
                     {
                        "operation": {
                           "methods": [
                              "POST"
                           ],
                           "paths": [
                              "/api/metrikker/vulnerabilities/trends/counts",
                              "/api/securityChampion/workMail",
                              "/api/scannerData",
                              "/api/scannerData/trends",
                              "/api/metrikker/ros-status",
                              "/api/metrikker/ros-status/v2",
                              "/api/dynamicScan/",
                              "/api/dynamicScan/*",
                              "/api/dynamicScan/**",
                              "/api/oppdateringer/alertsMetadata/accept"
                           ]
                        }
                     }
                  ],
                  "when": [
                     {
                        "key": "request.auth.claims[iss]",
                        "values": [
                           "https://login.microsoftonline.com/7f74c8a2-43ce-46b2-b0e8-b6306cba73a3/v2.0"
                        ]
                     }
                  ]
               },
               {
                  "from": [
                     {
                        "source": {
                           "principals": [
                              "cluster.local/ns/sikkerhetsmetrikker-main/sa/nightly-update-skipjob"
                           ]
                        }
                     }
                  ],
                  "to": [
                     {
                        "operation": {
                           "methods": [
                              "POST",
                              "GET"
                           ],
                           "paths": [
                              "/api/oppdateringer*"
                           ]
                        }
                     }
                  ]
               },
               {
                  "from": [
                     {
                        "source": {
                           "principals": [
                              "cluster.local/ns/sikkerhetsmetrikker-main/sa/nightly-update-skipjob"
                           ]
                        }
                     }
                  ],
                  "to": [
                     {
                        "operation": {
                           "methods": [
                              "POST"
                           ],
                           "paths": [
                              "/api/slack/nightlyUpdate",
                              "/api/slack/slackNotifications"
                           ]
                        }
                     }
                  ]
               },
               {
                  "from": [
                     {
                        "source": {
                           "principals": [
                              "cluster.local/ns/backstage/sa/backstage"
                           ]
                        }
                     }
                  ],
                  "to": [
                     {
                        "operation": {
                           "methods": [
                              "PUT"
                           ],
                           "paths": [
                              "/api/slack/configure-notifications"
                           ]
                        }
                     }
                  ]
               },
               {
                  "to": [
                     {
                        "operation": {
                           "methods": [
                              "POST"
                           ],
                           "paths": [
                              "/api/backstage/catalogInfo"
                           ]
                        }
                     }
                  ]
               },
               {
                  "to": [
                     {
                        "operation": {
                           "methods": [
                              "GET"
                           ],
                           "paths": [
                              "/api/public/metrikker/avdeling"
                           ]
                        }
                     }
                  ]
               }
            ],
            "selector": {
               "matchLabels": {
                  "app": "sikkerhetsmetrikker"
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
            "name": "smapi-db-user-password"
         },
         "spec": {
            "data": [
               {
                  "remoteRef": {
                     "key": "cloudsql-smapi-db-smapiapplication",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "DB_PASSWORD"
               },
               {
                  "remoteRef": {
                     "key": "cloudsql-smapi-db-instance",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "DB_HOST"
               }
            ],
            "refreshInterval": "1h0m0s",
            "secretStoreRef": {
               "kind": "SecretStore",
               "name": "smapi-gsm"
            },
            "target": {
               "name": "smapi-db-user-password"
            }
         }
      },
      {
         "apiVersion": "external-secrets.io/v1",
         "kind": "SecretStore",
         "metadata": {
            "name": "smapi-gsm"
         },
         "spec": {
            "provider": {
               "gcpsm": {
                  "projectID": "skvis-dev-5aa6"
               }
            }
         }
      }
   ],
   "kind": "List"
}