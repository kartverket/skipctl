function(
  env,
  version,
  name='backend',
  clientId,
  port,
  ingress,
) [
  {
    apiVersion: 'skiperator.kartverket.no/v1alpha1',
    kind: 'Application',
    metadata: {
      name: name,
    },
    variant: ingress,
    spec: {
      image: version,
      port: port,
      liveness: {
        path: '/health',
        port: port,
      },
      readiness: {
        path: '/health',
        port: port,
        variant: ingress,
      },
      resources: {
        requests: {
          cpu: '25m',
          memory: '256Mi',
        },
      },
      env: [
        {
          name: 'clientId',
          value: clientId,
        },

        {
          name: 'skipEnv',
          value: env,
        },
      ],
    },
  },
]
