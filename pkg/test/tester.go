package test

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	api "github.com/kartverket/skipctl/pkg/api/v1"
	"github.com/kartverket/skipctl/pkg/auth"
	"github.com/kartverket/skipctl/pkg/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/durationpb"
)

const highestPort = 65535

type Tester struct {
	client api.DiagnosticServiceClient
	log    *slog.Logger
}

func NewTester(_ context.Context, serverAddr string, useTLS bool) (*Tester, error) {
	var opts []grpc.DialOption

	adcCreds, err := auth.NewADCBackedRPCCredentials()
	if err != nil {
		return nil, err
	}

	opts = append(opts, grpc.WithPerRPCCredentials(adcCreds))

	if !useTLS {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(nil)))
	}

	conn, err := grpc.NewClient(serverAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc connection: %w", err)
	}

	return &Tester{
		client: api.NewDiagnosticServiceClient(conn),
		log:    logging.Logger().With("apiServer", serverAddr),
	}, nil
}

func (t *Tester) Ping(ctx context.Context, hostname string, count int32, timeout time.Duration) (*api.PingResponse, error) {
	t.log.InfoContext(ctx, "starting ping", "hostname", hostname, "count", count)
	res, err := t.client.Ping(ctx, &api.PingRequest{
		Host:    hostname,
		Count:   count,
		Timeout: durationpb.New(timeout),
	})

	return res, err
}

func (t *Tester) PortProbe(ctx context.Context, hostname string, port int32, timeout time.Duration) (*api.PortProbeResponse, error) {
	isValidPort := port >= 1 && port <= highestPort
	if !isValidPort {
		return nil, fmt.Errorf("invalid port: %d", port)
	}

	t.log.InfoContext(ctx, "starting port probe", "hostname", hostname, "port", port)
	res, err := t.client.PortProbe(context.Background(), &api.PortProbeRequest{
		Host:    hostname,
		Port:    port,
		Timeout: durationpb.New(timeout),
	})

	return res, err
}
