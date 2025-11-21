local argokit = import 'argokit/argokit.libsonnet';

// Complete example: Web application with database
argokit.Application {
  metadata: {
    name: 'web-app-example',
    namespace: 'production',
    labels: {
      'app.kubernetes.io/name': 'web-app',
      'app.kubernetes.io/version': '2.1.0',
      'app.kubernetes.io/component': 'frontend',
    },
  },
  spec: {
    containers: [
      {
        name: 'web',
        image: 'myregistry/web-app:2.1.0',
        ports: [
          { containerPort: 8080, name: 'http' },
          { containerPort: 9090, name: 'metrics' },
        ],
        
        // Environment configuration
        env: [
          { name: 'PORT', value: '8080' },
          { name: 'LOG_LEVEL', value: 'info' },
          {
            name: 'DATABASE_URL',
            valueFrom: {
              secretKeyRef: {
                name: 'db-connection',
                key: 'url',
              },
            },
          },
        ],
        
        // Resource limits
        resources: {
          limits: {
            cpu: '1',
            memory: '1Gi',
          },
          requests: {
            cpu: '200m',
            memory: '256Mi',
          },
        },
        
        // Health checks
        livenessProbe: {
          httpGet: {
            path: '/health',
            port: 'http',
          },
          initialDelaySeconds: 30,
          periodSeconds: 10,
          timeoutSeconds: 5,
          failureThreshold: 3,
        },
        
        readinessProbe: {
          httpGet: {
            path: '/ready',
            port: 'http',
          },
          initialDelaySeconds: 10,
          periodSeconds: 5,
          timeoutSeconds: 3,
          failureThreshold: 3,
        },
        
        // Security
        securityContext: {
          runAsNonRoot: true,
          runAsUser: 1000,
          readOnlyRootFilesystem: true,
          allowPrivilegeEscalation: false,
          capabilities: {
            drop: ['ALL'],
          },
        },
        
        // Volume mounts
        volumeMounts: [
          {
            name: 'tmp',
            mountPath: '/tmp',
          },
          {
            name: 'cache',
            mountPath: '/app/cache',
          },
        ],
      },
    ],
    
    // Volumes
    volumes: [
      {
        name: 'tmp',
        emptyDir: {},
      },
      {
        name: 'cache',
        emptyDir: {
          sizeLimit: '1Gi',
        },
      },
    ],
    
    // Autoscaling
    replicas: {
      min: 3,
      max: 20,
      targetCPUUtilization: 70,
    },
    
    // Service
    service: {
      type: 'ClusterIP',
      ports: [
        {
          name: 'http',
          port: 80,
          targetPort: 'http',
        },
      ],
    },
    
    // Ingress
    ingress: {
      enabled: true,
      host: 'webapp.example.com',
      tls: true,
      annotations: {
        'cert-manager.io/cluster-issuer': 'letsencrypt-prod',
        'nginx.ingress.kubernetes.io/rate-limit': '100',
      },
    },
  },
}
