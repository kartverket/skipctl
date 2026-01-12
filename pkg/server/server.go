package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
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
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

var log *slog.Logger

// validateServerConfig validates the server configuration parameters
func validateServerConfig(idTokenOrg, projectID, location, tlsCertFile, tlsKeyFile string) error {
	if len(idTokenOrg) == 0 {
		return errors.New("missing ID token organization")
	}
	if len(projectID) == 0 {
		return errors.New("missing GCP project ID")
	}
	if len(location) == 0 {
		return errors.New("missing GCP location")
	}
	// Validate TLS configuration
	if (tlsCertFile != "" && tlsKeyFile == "") || (tlsCertFile == "" && tlsKeyFile != "") {
		return errors.New("both --tls-cert and --tls-key must be provided together")
	}
	return nil
}

// setupTLSCredentials configures TLS credentials if certificate and key files are provided
func setupTLSCredentials(tlsCertFile, tlsKeyFile string) (credentials.TransportCredentials, error) {
	if tlsCertFile == "" || tlsKeyFile == "" {
		log.Warn("TLS not configured - server running without encryption. Use --tls-cert and --tls-key for production")
		return nil, nil //nolint:nilnil // nil credentials is a valid response when TLS is not configured
	}

	cert, err := tls.LoadX509KeyPair(tlsCertFile, tlsKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}
	log.Info("TLS enabled for gRPC server", "cert", tlsCertFile)
	return credentials.NewTLS(tlsConfig), nil
}

// setupGRPCServer creates and configures the gRPC server with authentication, metrics, and TLS
func setupGRPCServer(idTokenOrg, tlsCertFile, tlsKeyFile string, reg *prometheus.Registry) (*grpc.Server, *grpcprom.ServerMetrics, error) {
	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)
	reg.MustRegister(srvMetrics)

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			auth.ValidADCTokenWithOrg(idTokenOrg),
			srvMetrics.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			auth.ValidADCTokenWithOrgStream(idTokenOrg),
			srvMetrics.StreamServerInterceptor(),
		),
	}

	tlsCreds, err := setupTLSCredentials(tlsCertFile, tlsKeyFile)
	if err != nil {
		return nil, nil, err
	}
	if tlsCreds != nil {
		opts = append(opts, grpc.Creds(tlsCreds))
	}

	grpcSrv := grpc.NewServer(opts...)
	srvMetrics.InitializeMetrics(grpcSrv)
	return grpcSrv, srvMetrics, nil
}

// registerServices registers the diagnostic and AI services with the gRPC server
func registerServices(ctx context.Context, grpcSrv *grpc.Server, reg *prometheus.Registry, timeout time.Duration, projectID, location string) (*AIService, error) {
	// Register diagnostic service (optional)
	ds, err := NewDiagnosticService(reg, timeout)
	if err != nil {
		log.WarnContext(ctx, "diagnostic service unavailable (requires elevated permissions)", "error", err)
		log.InfoContext(ctx, "continuing without diagnostic service - only AI service will be available")
	} else {
		api.RegisterDiagnosticServiceServer(grpcSrv, ds)
		log.InfoContext(ctx, "diagnostic service registered")
	}

	// Register AI service (required)
	aiService, err := NewAIService(ctx, reg, timeout, projectID, location)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI service: %w", err)
	}
	api.RegisterAIServiceServer(grpcSrv, aiService)
	log.InfoContext(ctx, "AI service registered")

	reflection.Register(grpcSrv)
	return aiService, nil
}

// Serve starts a new API server capable of performing various probes for clients.
func Serve(addr string, metricsAddr string, timeout time.Duration, idTokenOrg string, projectID string, location string, tlsCertFile string, tlsKeyFile string) error {
	// Basic validation
	if log == nil {
		log = logging.Logger()
	}

	if err := validateServerConfig(idTokenOrg, projectID, location, tlsCertFile, tlsKeyFile); err != nil {
		return err
	}

	// Setup metrics registry
	reg := prometheus.NewRegistry()

	// Setup gRPC server with authentication, metrics, and TLS
	grpcSrv, _, err := setupGRPCServer(idTokenOrg, tlsCertFile, tlsKeyFile, reg)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Register services
	aiService, err := registerServices(ctx, grpcSrv, reg, timeout, projectID, location)
	if err != nil {
		return err
	}

	// Binding
	g := &run.Group{}

	g.Add(func() error {
		lc := net.ListenConfig{}
		l, lerr := lc.Listen(ctx, "tcp", addr)
		if lerr != nil {
			return err
		}
		log.Info("gRPC server listening", "addr", l.Addr())
		return grpcSrv.Serve(l)
	}, func(_ error) {
		grpcSrv.GracefulStop()
		grpcSrv.Stop()
		// Close AI service resources when server stops
		if closeErr := aiService.Close(); closeErr != nil {
			log.Error("failed to close AI service", "error", closeErr)
		}
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

	g.Add(run.SignalHandler(ctx, syscall.SIGINT, syscall.SIGTERM))

	if gerr := g.Run(); gerr != nil {
		return fmt.Errorf("failed to run server: %w", gerr)
	}

	return nil
}
