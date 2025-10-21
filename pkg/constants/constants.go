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
	SelfTestTimeout          = 2 * time.Second
)

const (
	ManifestSuffixJsonnet = ".jsonnet"
	ManifestSuffixYaml    = ".yaml"
	ManifestSuffixYml     = ".yml"
)

const (
	DiffOutputJSON   = "json"
	DiffOutputPatch  = "patch"
	DiffOutputPretty = "pretty"
)

const (
	DiffVerbosityMinimal = "minimal"
	DiffVerbosityChunk   = "chunk"
	DiffVerbosityFull    = "full"
)
const (
	DefaultChunkSize = 3
)

const (
	Equals    = "Equals"
	Deletion  = "Deletion"
	Insertion = "Insertion"
)

var (
	ManifestSuffixes = []string{ManifestSuffixJsonnet, ManifestSuffixYaml, ManifestSuffixYml}
)

// Collector constants
const (
	// TODO: Change these two constants before go live
	PostHogURL             = "https://eu.i.posthog.com"
	PostHogProjectAPIToken = ""
	BatchSize              = 1
	IDDirName              = "skipctl"
	IDFileName             = "id"
	RawIDBytes             = 16 // 128-bit random ID
	MinExistingLen         = 32 // hex length for 16 bytes
	CIDefault              = "IamCI"
)
