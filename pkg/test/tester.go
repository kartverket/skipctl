package test

import (
	"context"
	"fmt"
	"time"

	"github.com/kartverket/skipctl/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/durationpb"
)

type Tester struct {
	client api.DiagnosticServiceClient
}

func NewTester(serverAddr string, useTLS bool) (*Tester, error) {
	var opts []grpc.DialOption
	if !useTLS {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	var conn *grpc.ClientConn
	conn, err := grpc.NewClient(serverAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc connection: %w", err)
	}

	return &Tester{
		client: api.NewDiagnosticServiceClient(conn),
	}, nil
}

func (t *Tester) Ping(ctx context.Context, hostname string, count int32, timeout time.Duration) (*api.PingResponse, error) {
	res, err := t.client.Ping(ctx, &api.PingRequest{
		Host:    hostname,
		Count:   count,
		Timeout: durationpb.New(timeout),
	})

	return res, err
}
