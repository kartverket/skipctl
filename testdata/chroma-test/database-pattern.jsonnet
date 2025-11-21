local argokit = import 'argokit/argokit.libsonnet';

// Example: Database StatefulSet
argokit.Application {
  metadata: {
    name: 'postgres-db',
    namespace: 'databases',
  },
  spec: {
    // StatefulSet pattern
    stateful: true,
    containers: [
      {
        name: 'postgres',
        image: 'postgres:15',
        ports: [
          { containerPort: 5432, name: 'postgresql' },
        ],
        env: [
          {
            name: 'POSTGRES_PASSWORD',
            valueFrom: {
              secretKeyRef: {
                name: 'postgres-credentials',
                key: 'password',
              },
            },
          },
          { name: 'POSTGRES_DB', value: 'myapp' },
          { name: 'POSTGRES_USER', value: 'myapp' },
        ],
        volumeMounts: [
          {
            name: 'data',
            mountPath: '/var/lib/postgresql/data',
          },
        ],
        resources: {
          limits: {
            cpu: '2',
            memory: '4Gi',
          },
          requests: {
            cpu: '1',
            memory: '2Gi',
          },
        },
        livenessProbe: {
          exec: {
            command: ['pg_isready', '-U', 'myapp'],
          },
          initialDelaySeconds: 30,
          periodSeconds: 10,
        },
        readinessProbe: {
          exec: {
            command: ['pg_isready', '-U', 'myapp'],
          },
          initialDelaySeconds: 5,
          periodSeconds: 5,
        },
      },
    ],
    volumeClaimTemplates: [
      {
        metadata: {
          name: 'data',
        },
        spec: {
          accessModes: ['ReadWriteOnce'],
          resources: {
            requests: {
              storage: '50Gi',
            },
          },
          storageClassName: 'fast-ssd',
        },
      },
    ],
    replicas: {
      min: 1,
      max: 1,  // Databases typically don't autoscale
    },
  },
}
