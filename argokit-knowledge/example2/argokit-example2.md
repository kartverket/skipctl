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


application.new(
  name='foo-frontend',
  image='foo-frontend:1.2.3',
  port=3000
)

// 1. legg til miljøvariabler
+ application.withEnvironmentVariable('ENVIRONMENT', 'dev')

// 2. legg til access policies
+ application.withOutboundPostgres(host='database-host.com', ip='10.0.0.1')
+ application.withInboundSkipApp(appname='foo-backend')

// 3. sett opp ingress
+ application.forHostnames('foo-service.cloud.com')

// 4. konfigurer opp replicas
+ application.withReplicas(initial=2, max=10)

// 5. legg til probes
+ application.withLiveness(healthProbe)
+ application.withReadiness(healthProbe)


## Renders to this kubernetes application manifest:

{
   "apiVersion": "v1",
   "items": [
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