# ArgoKit Best Practices

## Resource Management

### Always Set Resource Limits

Every container should have CPU and memory limits defined:

```jsonnet
resources: {
  limits: {
    cpu: '1',
    memory: '512Mi',
  },
  requests: {
    cpu: '100m',
    memory: '128Mi',
  },
}
```

**Why?**
- Prevents resource starvation
- Enables proper scheduling
- Protects other workloads

## Health Checks

### Liveness Probes

Check if the application is alive and should be restarted:

```jsonnet
livenessProbe: {
  httpGet: {
    path: '/healthz',
    port: 8080,
  },
  initialDelaySeconds: 30,
  periodSeconds: 10,
  timeoutSeconds: 5,
  failureThreshold: 3,
}
```

### Readiness Probes

Check if the application is ready to receive traffic:

```jsonnet
readinessProbe: {
  httpGet: {
    path: '/ready',
    port: 8080,
  },
  initialDelaySeconds: 5,
  periodSeconds: 5,
  timeoutSeconds: 3,
  failureThreshold: 3,
}
```

## Security

### Security Context

Always run as non-root user:

```jsonnet
securityContext: {
  runAsNonRoot: true,
  runAsUser: 1000,
  readOnlyRootFilesystem: true,
  allowPrivilegeEscalation: false,
  capabilities: {
    drop: ['ALL'],
  },
}
```

### Secrets Management

Never hardcode secrets:

```jsonnet
env: [
  {
    name: 'DATABASE_PASSWORD',
    valueFrom: {
      secretKeyRef: {
        name: 'db-credentials',
        key: 'password',
      },
    },
  },
]
```

## Autoscaling

### Horizontal Pod Autoscaler

```jsonnet
replicas: {
  min: 2,
  max: 10,
  targetCPUUtilization: 70,
}
```

**Guidelines:**
- min >= 2 for production (high availability)
- max based on load testing
- targetCPU: 70-80% is typical

## Labels and Annotations

### Standard Labels

```jsonnet
metadata: {
  labels: {
    'app.kubernetes.io/name': 'myapp',
    'app.kubernetes.io/version': '1.0.0',
    'app.kubernetes.io/component': 'backend',
    'app.kubernetes.io/part-of': 'myplatform',
    'app.kubernetes.io/managed-by': 'argokit',
  },
}
```

## Environment Variables

### Configuration Precedence

1. Secrets (highest priority)
2. ConfigMaps
3. Default values in code

```jsonnet
env: [
  // From secret
  {
    name: 'API_KEY',
    valueFrom: {
      secretKeyRef: {
        name: 'api-credentials',
        key: 'key',
      },
    },
  },
  // From configmap
  {
    name: 'LOG_LEVEL',
    valueFrom: {
      configMapKeyRef: {
        name: 'app-config',
        key: 'logLevel',
      },
    },
  },
  // Direct value
  {
    name: 'FEATURE_FLAG',
    value: 'enabled',
  },
]
```

## Networking

### Service Definition

```jsonnet
service: {
  type: 'ClusterIP',
  ports: [
    {
      name: 'http',
      port: 80,
      targetPort: 8080,
    },
  ],
}
```

### Ingress

```jsonnet
ingress: {
  enabled: true,
  host: 'myapp.example.com',
  tls: true,
  annotations: {
    'cert-manager.io/cluster-issuer': 'letsencrypt-prod',
  },
}
```

## Persistent Storage

### Volume Claims

```jsonnet
volumeClaimTemplates: [
  {
    metadata: {
      name: 'data',
    },
    spec: {
      accessModes: ['ReadWriteOnce'],
      resources: {
        requests: {
          storage: '10Gi',
        },
      },
      storageClassName: 'fast-ssd',
    },
  },
]
```

## Monitoring

### Prometheus Annotations

```jsonnet
metadata: {
  annotations: {
    'prometheus.io/scrape': 'true',
    'prometheus.io/port': '9090',
    'prometheus.io/path': '/metrics',
  },
}
```
