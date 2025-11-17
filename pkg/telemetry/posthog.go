package telemetry

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/posthog/posthog-go"
)

var (
	trueVal = true
)

type Collector struct {
	log     *slog.Logger
	client  posthog.Client
	enabled bool
}

func (c *Collector) Enabled() bool {
	return c.enabled
}

type Options struct {
	Debug            bool
	DisableAnalytics bool
	GitVersion       string
	GitCommitHash    string
	Arch             string
	OS               string
}

func ConfigureCollector(opts Options) Collector {
	logger := logging.Logger().With("component", "telemetry/posthog")

	collector := Collector{log: logger}
	if len(PostHogProjectAPIToken) == 0 {
		collector.enabled = false
		logger.Info("telemetry is disabled because no PostHog project API token set – normal for development builds")
		return collector
	}

	if opts.DisableAnalytics || len(PostHogProjectAPIToken) == 0 {
		collector.enabled = false
		logger.Info("telemetry is disabled")
		return collector
	}

	logger.Info("telemetry enabled, set DO_NOT_TRACK=true to disable")

	config := posthog.Config{
		Endpoint:               postHogURL,
		BatchSize:              batchSize,
		DisableGeoIP:           &trueVal,
		Logger:                 &posthogSlogAdapter{logger},
		DefaultEventProperties: defaultProps(opts),
		Verbose:                opts.Debug,
	}

	client, err := posthog.NewWithConfig(PostHogProjectAPIToken, config)
	if err != nil {
		logger.Error("telemetry disabled: failed to initialize client", "error", err)
		collector.enabled = false
		return collector
	}
	collector.client = client
	collector.enabled = true
	return collector
}

func (c *Collector) Close() {
	if c.client != nil {
		if err := c.client.Close(); err != nil {
			c.log.Error("could not close PostHog client", "error", err)
		}
	}
}

// CaptureCommand sends a simple event. Safe to call even if telemetry is disabled.
func (c *Collector) CaptureCommand(command string, args []string, flags []string, runErr error) {
	// Skip if disabled or not initialized.
	if !c.enabled || c.client == nil || command == "" {
		return
	}

	var errMsg string
	if runErr != nil {
		errMsg = runErr.Error()
	}

	envKindVal := envKind()
	isCI := envKindVal == "ci"

	props := map[string]any{
		"command":   command,
		"args":      args,
		"flags":     flags,
		"had_error": runErr != nil,
		"error_msg": errMsg,
		"env_kind":  envKindVal,
		"time":      time.Now().UTC(),
		"$ip":       "0", // explicit neutral IP
	}

	var ciRepo string
	// Add GitHub Actions metadata as separate properties for easier filtering
	if isCI {
		if repo := os.Getenv("GITHUB_REPOSITORY"); repo != "" {
			ciRepo = repo
			props["ci_repository"] = repo
			props["ci_org"] = extractOrg(repo)
		}
		if workflow := os.Getenv("GITHUB_WORKFLOW"); workflow != "" {
			props["ci_workflow"] = workflow
		}
		if ref := os.Getenv("GITHUB_REF"); ref != "" {
			props["ci_ref"] = ref
		}
	}

	distinctID, hErr := hostHash(isCI, ciRepo)
	if hErr != nil {
		c.log.Error("could not get anonymous identity", "error", hErr)
	}
	if err := c.client.Enqueue(posthog.Capture{
		DistinctId: distinctID,
		Event:      "command",
		Properties: props,
	}); err != nil {
		c.log.Error("failed to enqueue telemetry event", "error", err)
	}
}

// envKind returns the environment kind: "local" or "ci".
func envKind() string {
	if os.Getenv("CI") == "true" {
		return "ci"
	}
	return "local"
}

// extractOrg extracts the organization/owner from a "owner/repo" string.
func extractOrg(repoSlug string) string {
	parts := strings.SplitN(repoSlug, "/", repoSlugSplitLimit)
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return "unknown"
}

func readOrCreateLocalID(isCI bool) (string, error) {
	// If it runs from a Action, just hash the hostname
	if isCI {
		return ciDefault, nil
	}
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	if cacheDir == "" {
		return "", errors.New("empty cache dir")
	}
	appDir := filepath.Join(cacheDir, idDirName)
	if mkErr := os.MkdirAll(appDir, 0o700); mkErr != nil {
		return "", mkErr
	}
	idPath := filepath.Join(appDir, idFileName)
	if b, readErr := os.ReadFile(idPath); readErr == nil && len(b) >= minExistingLen {
		return string(b), nil
	}
	buf := make([]byte, rawIDBytes)
	if _, genErr := rand.Read(buf); genErr != nil {
		return "", genErr
	}
	hexID := hex.EncodeToString(buf)
	if writeErr := os.WriteFile(idPath, []byte(hexID), 0o600); writeErr != nil {
		return "", writeErr
	}
	return hexID, nil
}

func hostHash(isCI bool, ciRepo string) (string, error) {
	id, err := readOrCreateLocalID(isCI)
	if err != nil {
		return "unknown", fmt.Errorf("failed to get local ID: %w", err)
	}

	// For CI: include repository in hash to distinguish between different repos
	if isCI && ciRepo != "" {
		id = id + ":" + ciRepo
	}

	h := sha256.Sum256([]byte(id))
	return hex.EncodeToString(h[:]), nil
}

func defaultProps(opts Options) posthog.Properties {
	props := make(posthog.Properties)

	props.Set("debug_mode", opts.Debug)
	props.Set("app_version", opts.GitVersion)
	props.Set("app_git_commit", opts.GitCommitHash)
	props.Set("user_os", runtime.GOOS)
	props.Set("user_arch", runtime.GOARCH)

	return props
}
