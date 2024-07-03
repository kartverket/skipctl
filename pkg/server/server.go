package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	api "github.com/kartverket/skipctl/pkg/api/v1"
	"github.com/kartverket/skipctl/pkg/auth"
	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var log *slog.Logger

// Serve starts a new API server capable of performing various probes for clients.
func Serve(addr string, metricsAddr string, timeout time.Duration, idTokenOrg string) error {
	// Basic validation
	if log == nil {
		log = logging.Logger()
	}

	if len(idTokenOrg) == 0 {
		return errors.New("missing ID token organization")
	}

	// Metrics
	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)
	reg := prometheus.NewRegistry()
	reg.MustRegister(srvMetrics)

	// gRPC
	opts := []grpc.ServerOption{grpc.ChainUnaryInterceptor(
		auth.ValidADCTokenWithOrg(idTokenOrg),
		srvMetrics.UnaryServerInterceptor()),
	}

	grpcSrv := grpc.NewServer(opts...)
	srvMetrics.InitializeMetrics(grpcSrv)

	// Register actual services
	ds, err := NewDiagnosticService(reg, timeout)
	if err != nil {
		return err
	}
	api.RegisterDiagnosticServiceServer(grpcSrv, ds)

	reflection.Register(grpcSrv)

	// Binding
	g := &run.Group{}

	g.Add(func() error {
		l, lerr := net.Listen("tcp", addr)
		if lerr != nil {
			return err
		}
		log.Info("gRPC server listening", "addr", l.Addr())
		return grpcSrv.Serve(l)
	}, func(_ error) {
		grpcSrv.GracefulStop()
		grpcSrv.Stop()
	})

	httpSrv := &http.Server{Addr: metricsAddr, ReadHeaderTimeout: constants.HTTPReadHeaderTimeout}
	g.Add(func() error {
		m := http.NewServeMux()
		// Create HTTP handler for Prometheus metrics.
		m.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
		httpSrv.Handler = m
		log.Info("metrics server listening", "addr", httpSrv.Addr)
		return httpSrv.ListenAndServe()
	}, func(_ error) {
		if httpErr := httpSrv.Close(); httpErr != nil {
			log.Error("failed to stop metrics web server", "error", httpErr)
		}
	})

	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))

	if gerr := g.Run(); gerr != nil {
		log.Error("failed to run server", "error", gerr)
		os.Exit(1)
	}

	return nil
}
