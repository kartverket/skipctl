function(name, env, version) [
  {
    apiVersion: 'skiperator.kartverket.no/v1alpha1',
    kind: 'Application',
    metadata: {
      name: name,
    },
    spec: {
      image: version,
      port: 3000,
      ingresses: ['devex.demo-frontend-' + env + '.host.com'],
      accessPolicy: {
        outbound: {
          rules: [
            {
              application: 'demo-backend',
            },
          ],
        },
      },
      env: [
        {
          name: 'VITE_AUTHORITY',
          value: 'https://login.microsoftonline.com/1234-5678/v2.0',
        },
        {
          name: 'VITE_FRONTEND_URL',
          value: 'https://devex.atgcp1-' + env + '.host.com',
        },
        {
          name: 'VITE_CLIENT_ID',
          value: if env == 'dev' then 'abcd-efgh-1234-5678'
          else if env == 'prod' then 'xyz-123-asdf',
        },
        {
          name: 'VITE_BACKEND_URL',
          value: '/api',
        },
      ],
    },
  },
]