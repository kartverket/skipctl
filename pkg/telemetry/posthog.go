package telemetry

import (
	"log"
	"os"
	"time"

	"github.com/posthog/posthog-go"
)

var client posthog.Client
var enabled bool

func Init(disabled bool) {
	if disabled {
		return
	}
	key := os.Getenv("POSTHOG_API_KEY")
	if key == "" {
		return
	}
	c, err := posthog.NewWithConfig(key, posthog.Config{
		Endpoint:  "https://eu.i.posthog.com",
		BatchSize: 1,
	})
	if err != nil {
		return
	}
	client = c
	enabled = true
}

func Enabled() bool { return enabled && client != nil }

func Close() {
	if client != nil {
		client.Close()
	}
}

// CaptureCommand sends a simple event. Safe to call even if disabled.
func CaptureCommand(command string, args []string, err error) {
	if !Enabled() {
		return
	}
	if command == "" { // skip root-only invocations
		return
	}
	var errMsg string
	var hadError bool
	if err != nil {
		hadError = true
		errMsg = err.Error()
	}
	props := map[string]any{
		"command":   command,
		"args":      args,
		"had_error": hadError,
		"error_msg": errMsg,
		"env_kind":  envKind(),
		"time":      time.Now().UTC(),
		"$ip":       "0", // explicit neutral IP

	}
	errEq := client.Enqueue(posthog.Capture{
		DistinctId: "test",
		Event:      "command",
		Properties: props,
	})
	if errEq != nil {
		log.Printf("Failed to enqueue telemetry event: %v", err)
	}
}

func envKind() string {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return "ci"
	}
	return "local"
}
