{
    apiVersion: 'skiperator.kartverket.no/v1alpha1',
    kind: 'Application',
    metadata: {
      name: 'frisk-backend',
    },
    spec: {
        container: { name: 'frisk-backend',
        dogs: 'are great'
        },
      image: '1',
      liveness: {
        path: '/health',
      },
      readiness: {
      },
      ingresses: [''],
      resources: {
        requests: {
          cpu: '25m',
          memory: '256Mi',
        },
      },
      env: [
        {
          name: 'clientId',
          value: 'dev',
        },
        {
          name: 'tenantId',
          value: '',
        },
        {
          name: 'environment',
          value: 'production',
        },
        {
          name: '',
          value: 'dev',
        },
        {
          name: '',
          value: 'postgres',
        },
        {
          name: 'REGELRETT_URL',
          value: '0',
        },
        {
          name: 'DATABASE_USERNAME',
          value: 'admin',
        },
        {
          name: 'ALLOWED_CORS_HOSTS',
          value:'dev',
        },

      ],
      envFrom: [
        {
          secret: 'dev',
        },
      ],
      filesFrom: [
        {
          mountPath: '',
          secret: 'db-ssl-ca',
        },
      ],
      accessPolicy: {
        inbound: {
          rules: [
            {
              application: '',
            },
          ],
        },
        outbound: {
          rules: [
            {
              application: '',
            },
          ],
          external: [
            {
              host: '',
            },
            {
              host: '',
            },
            {
              host: '',
              ip: '10.10.10.0/24',
              ports: [
                {
                  name: '',
                  port: 0,
                  protocol: '',
                },
              ],
            },
          ],
        },
      },
    },
  }