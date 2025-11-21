function(
  env,
  name='demo-backend',
  version,
  gsmProjectId,
  secretStoreName='devex-demo-gsm',
  esoName='devex-demo-secrets',
  kubernetesSecretEntraIdSecretName='DE-entraid-secret',
  cloudSqlConfig,
  dbUser,
  replyUrl,
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
      ingresses: ['backend.atgcp1-' + env + '.host.com'],
      resources: {
        requests: {
          cpu: '25m',
          memory: '256Mi',
        },
      },
      gcp: {
        cloudSqlProxy: cloudSqlConfig,
      },
      envFrom: [
        {
          secret: esoName,
        },
        {
          secret: kubernetesSecretEntraIdSecretName,
        },
      ],
      env: [
        {
          name: 'RR_SERVER_DOMAIN',
          value: 'demo.atgcp1-' + env + '.host.com',
        },
        {
          name: 'RR_SERVER_PROTOCOL',
          value: 'https',
        },
        {
          name: 'RR_SERVER_HTTP_PORT',
          value: '8080',
        },
        {
          name: 'RR_SERVER_ROOT_URL',
          value: 'https://demo.atgcp1-' + env + '.host.com',
        },
        {
          name: 'FRONTEND_URL_HOST',
          value: 'demo.atgcp1-' + env + '.host.com',
        },
        {
          name: 'TENANT_ID',
          value: 'f9f9f-abcd-1234-gagaga-lgtm-123',
        },
        {
          name: 'AUTH_PROVIDER_URL',
          value: 'https://demo.atgcp1-' + env + '.host.com/callback',
        },
        {
          name: 'DB_NAME',
          value: dbUser,
        },
        {
          name: 'DB_PASSWORD',
          value: '',
        },
        {
          name: 'FRISK_FRONTEND_URL_HOST',
          value: 'https://frisk.atgcp1-' + env + '.host.com',
        },
        {
          name: 'RR_OAUTH_TENANT_ID',
          value: 'f9f9f-abcd-1234-gagaga-lgtm-123',
        },
        {
          name: 'RR_DATABASE_HOST',
          value: 'localhost:5432',
        },
        {
          name: 'RR_DATABASE_NAME',
          value: 'demo',
        },
        {
          name: 'RR_DATABASE_USER',
          value: dbUser,
        },
        {
          name: 'RR_DATABASE_PASSWORD',
          value: '',
        },
        {
          name: 'RR_SERVER_ALLOWED_ORIGINS',
          value: 'frisk.atgcp1-' + env + '.host.com',
        },
        {
          name: 'RR_OAUTH_CLIENT_ID',
          valueFrom: {
            secretKeyRef: {
              name: kubernetesSecretEntraIdSecretName,
              key: 'RR_APP_CLIENT_ID',
            },
          },
        },
        {
          name: 'RR_OAUTH_CLIENT_SECRET',
          valueFrom: {
            secretKeyRef: {
              name: kubernetesSecretEntraIdSecretName,
              key: 'RR_APP_CLIENT_SECRET',
            },
          },
        },
      ],
      filesFrom: [
        {
          mountPath: '/etc/demo/provisioning/schemasources',
          secret: 'provisioning-file',
        },
      ],
      accessPolicy: {
        inbound: {
          rules: [
            {
              application: 'demo-frontend',
            },
            {
              application: 'frisk-backend',
            },
          ],
        },
        outbound: {
          external: [
            {
              host: 'api.airtable.com',
            },
            {
              host: 'login.microsoftonline.com',
            },
            {
              host: 'graph.microsoft.com',
            },
          ],
        },
      },
    },
  },
  {
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
  },
  {
    apiVersion: 'external-secrets.io/v1',
    kind: 'ExternalSecret',
    metadata: {
      name: esoName,
    },
    spec: {
      data: [
        {
          remoteRef: {
            key: 'demo-airtable-token',
            metadataPolicy: 'None',
          },
          secretKey: 'AIRTABLE_ACCESS_TOKEN',
        },
        {
          remoteRef: {
            key: 'demo-airtable-token',
            metadataPolicy: 'None',
          },
          secretKey: 'DE_SCHEMA_DRIFTSKONTINUITET_AIRTABLE_ACCESS_TOKEN',
        },
        {
          remoteRef: {
            key: 'demo-superuser',
            metadataPolicy: 'None',
          },
          secretKey: 'DE_OAUTH_SUPER_USER_GROUP',
        },
        {
          remoteRef: {
            key: 'demo-airtable-token',
            metadataPolicy: 'None',
          },
          secretKey: 'DE_SCHEMA_SIKKERHETSKONTROLLER_AIRTABLE_ACCESS_TOKEN',
        },
        {
          remoteRef: {
            key: 'demo-airtable-token',
            metadataPolicy: 'None',
          },
          secretKey: 'DE_AIRTABLE_ACCESS_TOKEN',
        },
      ],
      refreshInterval: '1h',
      secretStoreRef: {
        kind: 'SecretStore',
        name: secretStoreName,
      },
      target: {
        name: esoName,
      },
    },
  },
  {
    apiVersion: 'external-secrets.io/v1',
    kind: 'ExternalSecret',
    metadata: {
      name: 'provisioning-file',
    },
    spec: {
      data: [
        {
          remoteRef: {
            key: 'demo-defaults',
            metadataPolicy: 'None',
          },
          secretKey: 'defaults.yaml',
        },
      ],
      refreshInterval: '1h',
      secretStoreRef: {
        kind: 'SecretStore',
        name: secretStoreName,
      },
      target: {
        name: 'provisioning-file',
      },
    },
  },
  {
    apiVersion: 'external-secrets.io/v1',
    kind: 'SecretStore',
    metadata: {
      name: secretStoreName,
    },
    spec: {
      provider: {
        gcpsm: {
          projectID: gsmProjectId,
        },
      },
    },
  },
  {
    apiVersion: 'nais.io/v1',
    kind: 'AzureAdApplication',
    metadata: {
      name: 'demo-service-entraid',
      namespace: 'demo-main',
    },
    spec: {
      claims: {
        groups: [{ id: 'adsfdsa-gdadf-12346g-dadhjj' }],  //legger til kartverket i groups, for å assigne alle i kartverket til RR.
      },
      replyUrls: [
        {
          url: replyUrl,
        },
      ],
      preAuthorizedApplications: [
        {
          cluster: ' ',
          namespace: ' ',
          application: if env == 'dev' then '123456789-asdfghjkl-ddsddw'
          else if env == 'prod' then 'asdfg-12345-qwerty-45678',
        },
      ],
      secretName: kubernetesSecretEntraIdSecretName,
      secretKeyPrefix: 'RR_',
    },
  },
]