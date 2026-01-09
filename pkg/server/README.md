# Server Development Guide

This guide covers how to develop new gRPC services for skipctl.

## Prerequisites

Install the required tools for development:

```bash
# Install protobuf compiler
brew install protobuf

# Install Go protobuf plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Ensure Go binaries are in PATH
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Authenticate server

### Generate service account key

```bash
gcloud iam service-accounts keys create $HOME/.config/gcloud/skipctl-sa-key.json \
  --iam-account=skipctl@kv-spire-devex-ksde.iam.gserviceaccount.com
 ```

> **Security note:** Service account keys are powerful credentials that grant all permissions assigned to the service account. Treat this file as a secret:
>
> - Never commit `skipctl-sa-key.json` (or any service account key) to version control.
> - Add it to your `.gitignore` and do not share it over chat, email, or ticket systems.
> - Restrict file permissions (for example: `chmod 600 $HOME/.config/gcloud/skipctl-sa-key.json`).
> - Rotate keys regularly and delete unused keys from the service account.
> - For production and CI environments, prefer keyless authentication mechanisms (e.g., Workload Identity) instead of long‑lived keys.
## Running server locally

```bash
export GOOGLE_APPLICATION_CREDENTIALS=~/.config/gcloud/skipctl-sa-key.json && ./skipctl serve --gcp-project-id=kv-spire-devex-ksde --gcp-location=europe-north1
```

## Creating a New Service

### 1. Define the Protocol Buffer Service

Create a `.proto` file in `proto/api/v1/`:

```protobuf
syntax = "proto3";

package api.v1;

option go_package = "github.com/kartverket/skipctl/pkg/api/v1;apiv1";

service YourService {
  rpc YourMethod(YourRequest) returns (YourResponse);
}

message YourRequest {
  string field = 1;
}

message YourResponse {
  string result = 1;
}
```

### 2. Generate Go Code

Run the protobuf compiler:

```bash
protoc --go_out=./pkg --go-grpc_out=./pkg \
  --go_opt=paths=source_relative \
  --go-grpc_opt=paths=source_relative \
  proto/api/v1/your_service.proto
```

This generates:
- `pkg/api/v1/your_service.pb.go` (message definitions)
- `pkg/api/v1/your_service_grpc.pb.go` (service interface)

### 3. Implement the Service

Create a service file in `pkg/server/`:

```go
package server

import (
    "context"
    slogcontext "github.com/PumpkinSeed/slog-context"
    api "github.com/kartverket/skipctl/pkg/api/v1"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    yourServiceProcessed prometheus.Counter
    yourServiceOK        prometheus.Counter
    yourServiceFailed    prometheus.Counter
)

type YourService struct {
    api.UnimplementedYourServiceServer
    globalTimeout time.Duration
}

func NewYourService(reg *prometheus.Registry, globalTimeout time.Duration) (*YourService, error) {
    defineYourServiceMetrics(reg)
    
    return &YourService{
        globalTimeout: globalTimeout,
    }, nil
}

func defineYourServiceMetrics(reg *prometheus.Registry) {
    yourServiceProcessed = promauto.With(reg).NewCounter(prometheus.CounterOpts{
        Name: "your_service_processed_total",
        Help: "The total number of processed requests",
    })
    // Add more metrics as needed
}

func (s *YourService) YourMethod(ctx context.Context, req *api.YourRequest) (*api.YourResponse, error) {
    reqCtx := slogcontext.WithValue(ctx, "req", req)
    defer yourServiceProcessed.Inc()
    
    log.InfoContext(reqCtx, "received request")
    
    netCtx, cancel := globalTimeoutContext(reqCtx, s.globalTimeout)
    defer cancel()
    
    // Your implementation here
    
    return &api.YourResponse{
        Result: "success",
    }, nil
}
```

### 4. Register the Service

Add your service to `pkg/server/server.go`:

```go
// Register actual services
ds, err := NewDiagnosticService(reg, timeout)
if err != nil {
    return err
}
api.RegisterDiagnosticServiceServer(grpcSrv, ds)

// Register your new service
yourService, err := NewYourService(reg, timeout)
if err != nil {
    return fmt.Errorf("failed to create your service: %w", err)
}
api.RegisterYourServiceServer(grpcSrv, yourService)

reflection.Register(grpcSrv)
```

If your service needs additional dependencies (like external API clients), add them as parameters to the `Serve` function:

```go
func Serve(addr string, metricsAddr string, timeout time.Duration, idTokenOrg string, yourParam string) error {
    // ...existing validation...
    
    if len(yourParam) == 0 {
        return errors.New("missing your parameter")
    }
    
    // ...rest of function...
}
```

Then pass the parameters to your service constructor:

```go
yourService, err := NewYourService(reg, timeout, yourParam)
```

## Service Patterns

### Standard Patterns

- **Logging**: Use `slogcontext` for structured logging with request context
- **Metrics**: Define Prometheus metrics for monitoring
- **Timeouts**: Use `globalTimeoutContext` for consistent timeout handling
- **Error Handling**: Return gRPC errors with appropriate status codes

### Streaming Services

For streaming RPCs (like `AIService.AnalyzeFile`):

```go
func (s *YourService) StreamMethod(stream api.YourService_StreamMethodServer) error {
    ctx := stream.Context()
    
    // Receive data
    for {
        req, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
        // Process req
    }
    
    // Send response
    return stream.SendAndClose(&api.Response{})
}
```

## Example Services

- **DiagnosticService** (`diagnostic.go`): Simple request/response pattern with network operations
- **AIService** (`ai_service.go`): Streaming pattern with external API integration

## Testing

Create tests in `*_test.go` files:

```go
func TestYourMethod(t *testing.T) {
    reg := prometheus.NewRegistry()
    service, err := NewYourService(reg, 30*time.Second)
    require.NoError(t, err)
    
    resp, err := service.YourMethod(context.Background(), &api.YourRequest{
        Field: "test",
    })
    
    require.NoError(t, err)
    assert.Equal(t, "expected", resp.Result)
}
```

## Common Issues

- **Import paths**: Ensure `go_package` in `.proto` matches your module structure
- **Missing methods**: Embed `Unimplemented*Server` in your service struct
- **Metrics conflicts**: Use unique metric names with descriptive prefixes