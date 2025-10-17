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
	"time"

	"github.com/kartverket/skipctl/pkg/constants"
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

type Options struct {
	Debug            bool
	DisableAnalytics bool
	GitVersion       string
	GitCommitHash    string
	Arch             string
	OS               string
}

func ConfigureCollector(opts Options) *Collector {
	logger := logging.Logger().With("component", "telemetry/posthog")

	collector := &Collector{log: logger}
	if opts.DisableAnalytics {
		collector.enabled = false
		logger.Info("telemetry is disabled")
		return collector
	}

	logger.Info("telemetry enabled, set DO_NOT_TRACK=true to disable")

	config := posthog.Config{
		Endpoint:               constants.PostHogURL,
		BatchSize:              constants.BatchSize,
		DisableGeoIP:           &trueVal,
		Logger:                 &posthogSlogAdapter{logger},
		DefaultEventProperties: defaultProps(opts),
		Verbose:                true, // TODO: Remove
	}
	// NOTE: This is only here until we go live
	apiToken := os.Getenv("POSTHOG_API_KEY")
	if apiToken == "" {
		logger.Warn("POSTHOG_API_KEY environment variable not set, telemetry disabled")
		collector.enabled = false
		return collector
	}
	client, err := posthog.NewWithConfig(apiToken, config)
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
		err := c.client.Close()
		if err != nil {
			c.log.Error("could not close posthog client", "error", err)
			return
		}
	}
}

// CaptureCommand sends a simple event. Safe to call even if disabled.
func (c *Collector) CaptureCommand(command string, args []string, flags []string, runErr error) {
	// Skip if disabled or not initialized.
	if !c.enabled || c.client == nil || command == "" {
		return
	}

	var errMsg string
	if runErr != nil {
		errMsg = runErr.Error()
	}
	props := map[string]any{
		"command":   command,
		"args":      args,
		"flags":     flags,
		"had_error": runErr != nil,
		"error_msg": errMsg,
		"env_kind":  envKind(),
		"time":      time.Now().UTC(),
		"$ip":       "0", // explicit neutral IP
	}
	isCI := envKind() == "ci"
	distinctID, hErr := hostHash(isCI)
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

func envKind() string {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return "ci"
	}
	return "local"
}

func readOrCreateLocalID(isCI bool) (string, error) {
	// If it runs from a Action, just hash the hostname
	if isCI {
		return os.Hostname()
	}
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	if cacheDir == "" {
		return "", errors.New("empty cache dir")
	}
	appDir := filepath.Join(cacheDir, constants.IDDirName)
	if mkErr := os.MkdirAll(appDir, 0o700); mkErr != nil {
		return "", mkErr
	}
	idPath := filepath.Join(appDir, constants.IDFileName)
	if b, readErr := os.ReadFile(idPath); readErr == nil && len(b) >= constants.MinExistingLen {
		return string(b), nil
	}
	buf := make([]byte, constants.RawIDBytes)
	if _, genErr := rand.Read(buf); genErr != nil {
		return "", genErr
	}
	hexID := hex.EncodeToString(buf)
	if writeErr := os.WriteFile(idPath, []byte(hexID), 0o600); writeErr != nil {
		return "", writeErr
	}
	return hexID, nil
}

func hostHash(isCI bool) (string, error) {
	id, err := readOrCreateLocalID(isCI)
	if err != nil {
		return "unknown", fmt.Errorf("failed to get local ID: %w", err)
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
