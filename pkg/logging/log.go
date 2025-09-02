package logging

import (
	"log/slog"
	"os"
	"sync"

	slogcontext "github.com/PumpkinSeed/slog-context"
	"github.com/pkg/errors"
)

var logger *slog.Logger
var rawLogger *slog.Logger
var leveler *slog.LevelVar

var lock sync.Mutex

func init() {
	rawLogger = slog.New(&rawHandler{})
	leveler = new(slog.LevelVar)
	leveler.Set(slog.LevelInfo)
	// This is the default before anyone calls this function
	ConfigureLogging("json", false)
}

func ConfigureLogging(mode string, isDebug bool) *slog.Logger {
	parsedMode, err := parseOutputMode(mode)
	if err != nil {
		panic(err)
	}

	if isDebug {
		leveler.Set(slog.LevelDebug)
	}

	opts := &slog.HandlerOptions{
		Level:     leveler,
		AddSource: isDebug,
	}

	var h slog.Handler
	switch parsedMode {
	case OutputModeJSON:
		h = slog.NewJSONHandler(os.Stderr, opts)
	case OutputModeText:
		h = slog.NewTextHandler(os.Stderr, opts)
	default:
		panic(errors.Errorf("invalid output option: %v", parsedMode))
	}

	// slog-context outputs key-values found in the context to the log output
	ctxHandler := slogcontext.NewHandler(h)

	lock.Lock()
	defer lock.Unlock()
	logger = slog.New(ctxHandler)
	return logger
}

func Logger() *slog.Logger {
	if logger == nil {
		panic("logger not initialized")
	}

	return logger
}

// RawLogger returns a logger that just prints the message field as-is, without any extra formatting.
// This is useful for programs that need to output raw data.
func RawLogger() *slog.Logger {
	if rawLogger == nil {
		panic("logger not initialized")
	}

	return rawLogger
}
