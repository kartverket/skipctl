package logging

import (
	"log/slog"
	"os"
	"sync"

	slogcontext "github.com/PumpkinSeed/slog-context"
	"github.com/pkg/errors"
)

var logger *slog.Logger
var leveler *slog.LevelVar

var lock sync.Mutex

func init() {
	leveler = new(slog.LevelVar)
	leveler.Set(slog.LevelInfo)
	ConfigureLogging("json", false)
}

func ConfigureLogging(mode string, isDebug bool) *slog.Logger {
	parsedMode, err := ParseOutputMode(mode)
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
		h = slog.NewJSONHandler(os.Stdout, opts)
	case OutputModeText:
		h = slog.NewTextHandler(os.Stdout, opts)
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
