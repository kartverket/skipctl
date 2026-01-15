This file:
local argokit = import '../argokit/jsonnet/argokit.libsonnet';

function(image, db_host, db_ip) {
  apiVersion: 'skiperator.kartverket.no/v1alpha1',
  kind: 'Application',
  metadata: {
    name: 'eiendom',
  },
  spec: {
    image: image,
    port: 5000,
    envFrom: [
      { secret: 'eiendom-db-secret' },
    ],
    resources: {
      limits: {
        memory: '2G',
      },
      requests: {
        cpu: '100m',
        memory: '256M',
      },
    },
    prometheus: {
      path: '/eiendom/v1/metrics',
      port: 5000,
    },
    accessPolicy: {
      outbound: {
        external: [
          {
            host: db_host,
            ip: db_ip,
            ports: [
              {
                name: 'psql',
                protocol: 'TCP',
                port: 5432,
              },
            ],
          },
        ],
      },
    },
    liveness: {
      path: '/eiendom/v1/healthx',
      port: 5000,
    },
    readiness: {
      path: '/eiendom/v1/healthz',
      port: 5000,
    },
  },
}


Can be rewritten to argokit like this:

local argokit = import '../argokit/v2/jsonnet/argokit.libsonnet';
local application = argokit.appAndObjects.application;

function(image, db_host, db_ip)
  application.new(name='eiendom', image=image, port=5000)
  + application.withLiveness(
      application.probe(path='/eiendom/v1/healthx', port=5000)
    )
  + application.withReadiness(
      application.probe(path='/eiendom/v1/healthz', port=5000)
    )
  + application.withEnvironmentVariable(name='EIENDOM_INGRESS', value='eiendom.atkv3-dev.kartverket-intern.cloud')
  + application.withEnvironmentVariablesFromSecret('eiendom-db-secret')
  + application.forHostnames('eiendom.atkv3-dev.kartverket-intern.cloud')
  + application.withOutboundHttp(db_host, 'psql', 5432, 'TCP')
  + application.resources.withRequests(cpu='100m', memory='256M')
  + application.resources.withLimits(memory='2G')
  + {
    application+: {
      spec+: {
        prometheus: {
          path: '/eiendom/v1/metrics',
          port: 5000,
        },
      },
    },
  }