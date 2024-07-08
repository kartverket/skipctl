package server

import (
	context "context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/kartverket/skipctl/pkg/api"
	probing "github.com/prometheus-community/pro-bing"
	"google.golang.org/protobuf/types/known/durationpb"
)

// TODO: Metrics

const dnsResolveTimeout = 2 * time.Second

type DiagnosticService struct {
	api.UnimplementedDiagnosticServiceServer
	globalTimeout time.Duration
}

func (d DiagnosticService) Ping(ctx context.Context, req *api.PingRequest) (*api.PingResponse, error) {
	log.InfoContext(ctx, "received ping request")
	netCtx, cancel := globalTimeoutContext(ctx, d.globalTimeout)
	defer cancel()

	res, err := doPing(netCtx, req.GetHost(), int(req.GetCount()), protoDuration(req.GetTimeout()))
	if err != nil {
		log.WarnContext(ctx, "ping failed", "error", err)
		return nil, err
	}

	return res, nil
}

func (d DiagnosticService) PortProbe(ctx context.Context, req *api.PortProbeRequest) (*api.PortProbeResponse, error) {
	netCtx, cancel := globalTimeoutContext(ctx, d.globalTimeout)
	defer cancel()

	log.InfoContext(ctx, "received port probe request")
	res, err := doProbe(netCtx, req.GetHost(), int(req.GetPort()), protoDuration(req.GetTimeout()))
	if err != nil {
		log.InfoContext(ctx, "port probe failed", "error", err)
	}
	return res, nil
}

func doPing(ctx context.Context, host string, count int, timeout *time.Duration) (*api.PingResponse, error) {
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
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	stats := p.Statistics()
	return &api.PingResponse{
		Pingable:             stats.PacketsRecv > 0,
		PacketsReceived:      int32(stats.PacketsRecv),
		PacketsSent:          int32(stats.PacketsSent),
		PacketLossPercentage: int32(stats.PacketLoss),
		PingedHost:           stats.IPAddr.String(),
		MinRtt:               durationpb.New(stats.MinRtt),
		MaxRtt:               durationpb.New(stats.MaxRtt),
		AvgRtt:               durationpb.New(stats.AvgRtt),
		StdDevRtt:            durationpb.New(stats.StdDevRtt),
	}, nil
}

func doProbe(ctx context.Context, host string, port int, timeout *time.Duration) (*api.PortProbeResponse, error) {
	var d net.Dialer
	if timeout != nil {
		d.Timeout = *timeout
	}
	var remoteAddr string
	conn, err := d.DialContext(ctx, "tcp4", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return &api.PortProbeResponse{
			Open:       false,
			AddrProbed: nil,
		}, err
	}
	defer conn.Close()
	remoteAddr = conn.RemoteAddr().String()

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
