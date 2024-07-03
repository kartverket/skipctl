package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	slogcontext "github.com/PumpkinSeed/slog-context"
	api "github.com/kartverket/skipctl/pkg/api/v1"
	probing "github.com/prometheus-community/pro-bing"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/protobuf/types/known/durationpb"
)

// TODO: Metrics

const dnsResolveTimeout = 2 * time.Second

var (
	pingProcessed prometheus.Counter
	pingOK        *prometheus.CounterVec
	pingFailed    *prometheus.CounterVec

	probeProcessed prometheus.Counter
	probeOK        *prometheus.CounterVec
	probeFailed    *prometheus.CounterVec
)

type DiagnosticService struct {
	api.UnimplementedDiagnosticServiceServer
	globalTimeout time.Duration
}

func NewDiagnosticService(reg *prometheus.Registry, globalTimeout time.Duration) (*DiagnosticService, error) {
	conn, err := net.Dial("ip4:icmp", "127.0.0.1")
	if err != nil {
		return nil, errors.New("unable to do raw ICMP sockets – missing permissions")
	}
	_ = conn.Close()
	defineMetrics(reg)

	return &DiagnosticService{
		globalTimeout: globalTimeout,
	}, nil
}

func defineMetrics(reg *prometheus.Registry) {
	pingProcessed = promauto.With(reg).NewCounter(prometheus.CounterOpts{
		Name: "pings_processed_total",
		Help: "The total number of processed pings",
	})
	pingOK = promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
		Name: "pings_ok_total",
		Help: "The total number of OK pings",
	}, []string{"hostname"})
	pingFailed = promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
		Name: "pings_failed_total",
		Help: "The total number of failed pings",
	}, []string{"hostname"})

	probeProcessed = promauto.With(reg).NewCounter(prometheus.CounterOpts{
		Name: "probes_processed_total",
		Help: "The total number of processed port probes",
	})
	probeOK = promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
		Name: "probes_ok_total",
		Help: "The total number of OK port probes",
	}, []string{"hostname", "port"})
	probeFailed = promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
		Name: "probes_failed_total",
		Help: "The total number of failed port probes",
	}, []string{"hostname", "port"})
}

func (d DiagnosticService) Ping(ctx context.Context, req *api.PingRequest) (*api.PingResponse, error) {
	reqCtx := slogcontext.WithValue(ctx, "req", req)

	log.InfoContext(reqCtx, "received ping request")
	netCtx, cancel := globalTimeoutContext(reqCtx, d.globalTimeout)
	defer cancel()

	res, err := doPing(netCtx, req.GetHost(), int(req.GetCount()), protoDuration(req.GetTimeout()))
	if err != nil {
		log.WarnContext(reqCtx, "ping failed", "error", err)
		return nil, err
	}

	return res, nil
}

func (d DiagnosticService) PortProbe(ctx context.Context, req *api.PortProbeRequest) (*api.PortProbeResponse, error) {
	reqCtx := slogcontext.WithValue(ctx, "req", req)
	netCtx, cancel := globalTimeoutContext(reqCtx, d.globalTimeout)
	defer cancel()

	log.InfoContext(reqCtx, "received port probe request")
	res, err := doProbe(netCtx, req.GetHost(), int(req.GetPort()), protoDuration(req.GetTimeout()))
	if err != nil {
		log.InfoContext(reqCtx, "port probe failed", "error", err)
	}
	return res, nil
}

func doPing(ctx context.Context, host string, count int, timeout *time.Duration) (*api.PingResponse, error) {
	defer pingProcessed.Inc()
	p, err := probing.NewPinger(host)
	if err != nil {
		return nil, err
	}
	// required to do proper ICMP ping
	p.SetPrivileged(true)
	p.Count = count
	p.ResolveTimeout = dnsResolveTimeout
	if timeout != nil {
		p.Timeout = *timeout
	}

	if err = p.RunWithContext(ctx); err != nil {
		pingFailed.With(prometheus.Labels{"hostname": host}).Inc()
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	stats := p.Statistics()
	r := &api.PingResponse{
		Pingable:             stats.PacketsRecv > 0,
		PacketsReceived:      int32(stats.PacketsRecv),
		PacketsSent:          int32(stats.PacketsSent),
		PacketLossPercentage: int32(stats.PacketLoss),
		PingedHost:           stats.IPAddr.String(),
		MinRtt:               durationpb.New(stats.MinRtt),
		MaxRtt:               durationpb.New(stats.MaxRtt),
		AvgRtt:               durationpb.New(stats.AvgRtt),
		StdDevRtt:            durationpb.New(stats.StdDevRtt),
	}

	if r.GetPingable() {
		pingOK.With(prometheus.Labels{"hostname": host}).Inc()
	} else {
		pingFailed.With(prometheus.Labels{"hostname": host}).Inc()
	}

	return r, nil
}

func doProbe(ctx context.Context, host string, port int, timeout *time.Duration) (*api.PortProbeResponse, error) {
	defer probeProcessed.Inc()
	var d net.Dialer
	if timeout != nil {
		d.Timeout = *timeout
	}
	var remoteAddr string
	conn, err := d.DialContext(ctx, "tcp4", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		probeFailed.With(prometheus.Labels{"hostname": host, "port": strconv.Itoa(port)}).Inc()
		return &api.PortProbeResponse{
			Open:       false,
			AddrProbed: nil,
		}, err
	}
	defer conn.Close()
	remoteAddr = conn.RemoteAddr().String()

	probeOK.With(prometheus.Labels{"hostname": host, "port": strconv.Itoa(port)}).Inc()
	return &api.PortProbeResponse{
		Open:       true,
		AddrProbed: &remoteAddr,
	}, nil
}

func globalTimeoutContext(parentContext context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeoutCause(parentContext, timeout, errors.New("skipctl max timeout reached"))
}

func protoDuration(pbDuration *durationpb.Duration) *time.Duration {
	if pbDuration == nil {
		return nil
	}
	d := pbDuration.AsDuration()
	return &d
}
