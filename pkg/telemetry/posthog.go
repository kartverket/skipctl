package telemetry

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/posthog/posthog-go"
)

type Collector struct {
	client  posthog.Client
	config  posthog.Config
	enabled bool
}

func ConfigureCollector(disabledAnalytics bool) *Collector {
	collector := &Collector{}
	if disabledAnalytics {
		collector.enabled = false
		return collector
	}
	key := os.Getenv("POSTHOG_API_KEY")
	if key == "" {
		log.Printf("telemetry disabled: missing POSTHOG_API_KEY")
		collector.enabled = false
		return collector
	}
	config := posthog.Config{
		Endpoint:     constants.PostHogURL,
		BatchSize:    constants.BatchSize,
		DisableGeoIP: &constants.DisableGeoIP,
	}
	client, err := posthog.NewWithConfig(key, config)
	if err != nil {
		log.Printf("telemetry disabled: failed to initialize client: %v", err)
		collector.enabled = false
		return collector
	}
	collector.client = client
	collector.config = config
	collector.enabled = true
	return collector
}

func (c *Collector) Close() {
	if c.client != nil {
		c.client.Close()
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
	distinctID := hostHash()
	if err := c.client.Enqueue(posthog.Capture{
		DistinctId: distinctID,
		Event:      "command",
		Properties: props,
	}); err != nil {
		log.Printf("Failed to enqueue telemetry event: %v", err)
	}
}

func envKind() string {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return "ci"
	}
	return "local"
}
func readOrCreateLocalID() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(cacheDir, "skipctl")
	if err := os.MkdirAll(appDir, 0o700); err != nil {
		return "", err
	}
	idPath := filepath.Join(appDir, "id")
	if b, err := os.ReadFile(idPath); err == nil && len(b) >= 32 {
		return string(b), nil
	}
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	hexID := hex.EncodeToString(buf)
	if err := os.WriteFile(idPath, []byte(hexID), 0o600); err != nil {
		return "", err
	}
	return hexID, nil
}

func hostHash() string {
	id, err := readOrCreateLocalID()
	if err != nil {
		log.Printf("Failed to get local ID: %v", err)
		return "unknown"
	}
	h := sha256.Sum256([]byte(id))
	return hex.EncodeToString(h[:])
}
