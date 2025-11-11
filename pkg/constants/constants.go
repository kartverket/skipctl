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
	ManifestKustomizeYaml = "kustomization.yaml"
	ManifestKustomizeYml  = "kustomization.yml"
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

// Commands we dont want to track
const (
	Completion = "completion"
)

var NotTrackableCommands = []string{Completion}
