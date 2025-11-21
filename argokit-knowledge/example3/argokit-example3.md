## This argokit v2:

local argokit = import '../../argokit/v2/jsonnet/argokit.libsonnet';
local application = argokit.appAndObjects.application;

local healthProbe = application.probe(
  path='/health',
  port=8080,
  failureThreshold=5,
  timeout=0,
  initialDelay=5
);

local secrets = [
  {
    fromSecret: 'db-password',
    toKey: 'DB_PASS',
  },
  {
    fromSecret: 'db-user',
    toKey: 'DB_USER',
  },
];

application.new(
  name='foo-frontend',
  image='foo-frontend:1.2.3',
  port=3000
)

+ application.withEnvironmentVariable('ENVIRONMENT', 'dev')
+ application.withOutboundPostgres(host='database-host.com', ip='10.0.0.1')
+ application.withInboundSkipApp(appname='foo-backend')
+ application.forHostnames('foo-service.cloud.com')
+ application.withReplicas(initial=2, max=10)
+ application.withLiveness(healthProbe)
+ application.withReadiness(healthProbe)

+ application.withAzureAdApplication(
  name='foo-ad',
  namespace='foo-team-main',
  secretPrefix='foosecrets',
)

+ application.withEnvironmentVariablesFromExternalSecret(
  name='database-secrets',
  secrets=secrets,
)

## Renders to this Kubernetes Application manifest:

{
   "apiVersion": "v1",
   "items": [
      {
         "apiVersion": "external-secrets.io/v1",
         "kind": "ExternalSecret",
         "metadata": {
            "name": "database-secrets"
         },
         "spec": {
            "data": [
               {
                  "remoteRef": {
                     "key": "db-password",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "DB_PASS"
               },
               {
                  "remoteRef": {
                     "key": "db-user",
                     "metadataPolicy": "None"
                  },
                  "secretKey": "DB_USER"
               }
            ],
            "refreshInterval": "1h",
            "secretStoreRef": {
               "kind": "SecretStore",
               "name": "gsm"
            },
            "target": {
               "name": "database-secrets"
            }
         }
      },
      {
         "apiVersion": "nais.io/v1",
         "kind": "AzureAdApplication",
         "metadata": {
            "name": "foo-ad",
            "namespace": "foo-team-main"
         },
         "spec": {
            "allowAllUsers": false,
            "replyUrls": [
               {
                  "url": "http://localhost/callback"
               }
            ],
            "secretName": "foosecrets-foo-ad"
         }
      },
      {
         "apiVersion": "skiperator.kartverket.no/v1alpha1",
         "kind": "Application",
         "metadata": {
            "name": "foo-frontend"
         },
         "spec": {
            "accessPolicy": {
               "inbound": {
                  "rules": [
                     {
                        "application": "foo-backend"
                     }
                  ]
               },
               "outbound": {
                  "external": [
                     {
                        "host": "database-host.com",
                        "ip": "10.0.0.1",
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
                     }
                  ]
               }
            },
            "env": [
               {
                  "name": "ENVIRONMENT",
                  "value": "dev"
               }
            ],
            "envFrom": [
               {
                  "secret": "foosecrets-foo-ad"
               },
               {
                  "secret": "database-secrets"
               }
            ],
            "image": "foo-frontend:1.2.3",
            "ingresses": [
               "foo-service.cloud.com"
            ],
            "liveness": {
               "failureThreshold": 5,
               "initialDelay": 5,
               "path": "/health",
               "port": 8080,
               "timeout": 0
            },
            "port": 3000,
            "readiness": {
               "failureThreshold": 5,
               "initialDelay": 5,
               "path": "/health",
               "port": 8080,
               "timeout": 0
            },
            "replicas": {
               "max": 10,
               "min": 2,
               "targetCpuUtilization": 80
            }
         }
      }
   ],
   "kind": "List"
}