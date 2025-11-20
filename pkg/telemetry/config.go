package telemetry

const (
	postHogURL         = "https://ph.kartverket.no"
	batchSize          = 1
	idDirName          = "skipctl"
	idFileName         = "id"
	rawIDBytes         = 16 // 128-bit random ID
	minExistingLen     = 32 // hex length for 16 bytes
	ciDefault          = "running-in-ci"
	repoSlugSplitLimit = 2 // Split "owner/repo" into max 2 parts
)

var PostHogProjectAPIToken string
