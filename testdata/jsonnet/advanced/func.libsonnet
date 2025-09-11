function(
  env,
  version,
  name='backend',
  clientId,
) [
  {
    apiVersion: 'skiperator.kartverket.no/v1alpha1',
    kind: 'Application',
    metadata: {
      name: name,
    },
    spec: {
      image: version,
      port: 8080,
      liveness: {
        path: '/health',
        port: 8080,
      },
      readiness: {
        path: '/health',
        port: 8080,
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
