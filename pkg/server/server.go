package server

import (
	"errors"
	"log/slog"
	"net"
	"os"
	"time"

	api "github.com/kartverket/skipctl/pkg/api/v1"
	"github.com/kartverket/skipctl/pkg/auth"
	"github.com/kartverket/skipctl/pkg/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var log *slog.Logger

// TODO: check for privileges to do ICMP

// Serve starts a new API server capable of performing various probes for clients.
func Serve(addr string, timeout time.Duration, idTokenOrg string) error {
	if log == nil {
		log = logging.Logger()
	}

	if len(idTokenOrg) == 0 {
		return errors.New("missing ID token organization")
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
