package server

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/kartverket/skipctl/pkg/api"
	"github.com/kartverket/skipctl/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var log *slog.Logger

func init() {
	log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	auth.SetLogger(log)
}

// TODO: check for priviliges to do ICMP
func Serve(addr string, timeout time.Duration, idTokenOrg string) error {
	if len(idTokenOrg) == 0 {
		return fmt.Errorf("missing ID token organization")
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(auth.ValidADCTokenWithOrg(idTokenOrg)),
	}
	s := grpc.NewServer(opts...)

	// Register services
	api.RegisterDiagnosticServiceServer(s, &DiagnosticService{
		log:           log,
		globalTimeout: timeout,
	})

	reflection.Register(s)
	log.Info("gRPC server listening", "addr", l.Addr())
	if err = s.Serve(l); err != nil {
		log.Error("failed to serve", "error", err)
		os.Exit(1)
	}

	return nil
}
