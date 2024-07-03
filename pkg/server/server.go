package server

import (
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/kartverket/skipctl/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var log *slog.Logger

func init() {
	log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

// TODO: check for priviliges to do ICMP
func Serve(addr string, timeout time.Duration) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s := grpc.NewServer()

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
