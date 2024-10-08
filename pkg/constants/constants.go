package constants

import "time"

const (
	DefaultDiscoveryServer   = "_skipctl.kartverket-intern.cloud"
	DefaultTestTimeout       = 10 * time.Second
	DefaultServerTestTimeout = 1 * time.Minute
	DefaultPingCount         = 10
	DefaultGoogleOrgID       = "kartverket.no"
	DNSDiscoverTimeout       = 5 * time.Second
	HTTPReadHeaderTimeout    = 5 * time.Second
)
