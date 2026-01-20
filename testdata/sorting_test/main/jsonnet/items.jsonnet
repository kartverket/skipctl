{
  apiVersion: "v1",
  items: [
    {
      apiVersion: "apps/v1",
      kind: "Deployment",
      metadata: {
        name: "my-deployment",
        namespace: "default"
      },
      spec: {
        replicas: 1,
        selector: {
          matchLabels: {
            app: "myapp"
          }
        },
        template: {
          metadata: {
            labels: {
              app: "myapp"
            }
          },
          spec: {
            containers: [
              {
                name: "nginx",
                image: "nginx:latest"
              }
            ]
          }
        }
      }
    },
    {
      apiVersion: "v1",
      kind: "Service",
      metadata: {
        name: "my-service",
        namespace: "default"
      },
      spec: {
        ports: [
          {
            port: 80,
            targetPort: 8080
          }
        ],
        selector: {
          app: "myapp"
        }
      }
    }
  ],
  kind: "List"
}
